package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/gearbox/process"
	"github.com/cenkalti/backoff"
	"github.com/ghodss/yaml"
	"golang.org/x/net/context"
)

type command struct {
	Cmd string
}

func read(file string) (cmd []string, er error) {
	logger.Debugf("Reading command config %q", file)
	body, er := ioutil.ReadFile(file)
	if er != nil {
		return
	}

	c := new(command)
	if er = yaml.Unmarshal(body, c); er != nil {
		return
	}

	logger.Debugf("Raw command: %s", c.Cmd)
	cmd = strings.Fields(os.ExpandEnv(c.Cmd))
	return
}

func prepareApp(ctx context.Context) (app *process.Process, er error) {
	var (
		cmd []string
	)

	if cmd, er = read(*conf); er != nil {
		return
	}

	logger.Debugf("Looking for %q in PATH", cmd[0])
	if cmd[0], er = exec.LookPath(cmd[0]); er != nil {
		return
	}

	if app, er = process.New(*name, strings.Join(cmd, " "), stdout...); er != nil {
		return
	}

	if len(*env) > 0 {
		e := make([]string, 0, len(*env))
		for k, v := range *env {
			e = append(e, k+"="+v)
		}
		app.SetEnv(e)
	}

	app.SetDir(path.Join("/src", *name))
	app.SetUser(*uid, *gid)
	return
}

func run(app *process.Process, c context.Context, cancel context.CancelFunc) {
	var (
		failures = 0
		up       = time.NewTicker(*interval)
	)

	if er := app.Execute(c); er != nil {
		logger.Errorf("%v failed to execute: %v", app, er)
		return
	}

	for failures < 10 {
		select {
		case <-c.Done():
			return
		case <-app.Exited():
			if er := app.Execute(c); er != nil {
				logger.Errorf("%v failed to execute: %v", app, er)
				failures++
				time.Sleep(2 * time.Minute)
			}
		case t := <-up.C:
			logger.Infof("Updating %v at %v", *name, t.Format(time.Stamp))
			updated, er := update(c)
			if er != nil {
				logger.Errorf("Failed update: %v", er)
				continue
			}
			if updated {
				logger.Infof("Restarting %v", app)
				if er := kill(app, c); er != nil {
					logger.Errorf("Failed to kill %v: %v", app, er)
					failures++
				}
			}
		}
	}

	cancel()
}

func kill(app *process.Process, c context.Context) error {
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = 60 * time.Second

	notify := func(er error, to time.Duration) {
		logger.Warnf("Failed to kill %v (retry in %v): %v", app, to, er)
	}
	operation := func() error {
		t := time.NewTimer(4 * time.Second)
		defer t.Stop()

		if er := app.Process.Kill(); er != nil {
			return er
		}

		select {
		case <-c.Done():
			return nil
		case <-app.Exited():
			return nil
		case <-t.C:
			return errors.New("timeout")
		}
	}

	return backoff.RetryNotify(operation, exp, notify)
}

func update(c context.Context) (bool, error) {
	var (
		dir    = path.Join("/src", *name)
		remote = []string{"remote", "update", "-p"}
		merge  = []string{
			"merge",
			"--ff",
			"--strategy",
			"recursive",
			"-Xpatience",
			"-Xrenormalize",
			"@{u}",
		}
	)
	// git remote update -p
	rem, er := process.New(
		"git-remote-update",
		strings.Join(append([]string{git}, remote...), " "),
		stdout...,
	)
	if er != nil {
		return false, er
	}

	if er := rem.SetDir(dir).SetUser(*uid, *gid).Execute(c); er != nil {
		return false, er
	}
	<-rem.Exited()

	// git checkout branch
	co, er := process.New(
		fmt.Sprintf("git-checkout-%s", ref),
		strings.Join([]string{git, "checkout", ref}, " "),
		stdout...,
	)
	if er != nil {
		return false, er
	}

	if er := co.SetDir(dir).SetUser(*uid, *gid).Execute(c); er != nil {
		return false, er
	}
	<-co.Exited()

	// git merge
	me, er := process.New(
		"git-merge",
		strings.Join(append([]string{git}, merge...), " "),
		stdout...,
	)
	if er != nil {
		return false, er
	}

	if er := me.SetDir(dir).SetUser(*uid, *gid).Execute(c); er != nil {
		return false, er
	}
	<-me.Exited()

	// find sha
	head, er := getSHA()
	if er != nil {
		return false, er
	}
	return sha != head, nil
}

func clone(c context.Context) error {
	var (
		args = []string{"clone", "--recursive", "--single-branch", "--progress"}
	)
	logger.Infof("Cloning %q", *project)

	if er := os.Mkdir("/src", 0755); er != nil {
		return er
	}
	if er := os.Chown("/src", int(*uid), int(*gid)); er != nil {
		return er
	}

	loc := fmt.Sprintf("git://github.com/%s.git", *project)
	if *name == "" {
		*name = strings.ToLower(path.Base(*project))
	}
	dir := path.Join("/src", *name)
	if er := os.Setenv("APP_HOME", dir); er != nil {
		logger.Warnf("Unable to set APP_HOME env var: %v", er)
	}

	if *branch != "" {
		args = append(args, "--branch", *branch)
	}
	args = append(args, loc, dir)

	co, er := process.New(
		fmt.Sprintf("git-clone-%s", *name),
		strings.Join(append([]string{git}, args...), " "),
		stdout...,
	)
	if er != nil {
		return er
	}

	logger.Debugf("Executing %v", co)
	if er := co.SetDir("/src").SetUser(*uid, *gid).Execute(c); er != nil {
		return er
	}
	<-co.Exited()
	return os.Chdir(dir)
}

func getSHA() (string, error) {
	logger.Infof("Determining HEAD sha")
	b := new(bytes.Buffer)

	sh := exec.Command(git, "rev-parse", "HEAD")
	sh.Dir = path.Join("/src", *name)
	sh.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: *uid,
			Gid: *gid,
		},
	}
	sh.Stdout = b

	if er := sh.Run(); er != nil {
		return "", er
	}
	logger.Infof("HEAD sha: %s", b.String()[:10])
	return strings.TrimSpace(b.String()), nil
}

func getRef() (string, error) {
	logger.Infof("Determining current ref")
	b := new(bytes.Buffer)

	re := exec.Command(git, "rev-parse", "--abbrev-ref", "HEAD")
	re.Dir = path.Join("/src", *name)
	re.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: *uid,
			Gid: *gid,
		},
	}
	re.Stdout = b

	if er := re.Run(); er != nil {
		return "", er
	}
	r := strings.TrimSpace(b.String())
	logger.Infof("Current ref: %s", r)
	return r, nil
}
