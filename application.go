package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/spf13/afero"
)

var (
	fs afero.Fs
)

type App struct {
	version string
	restart bool
	command *exec.Cmd
	Name    string `json:"name,omitempty"`
	Dir     string `json:"directory,omitempty"`
	Cmd     string `json:"command,omitempty"`
}

func newApp(c string) *App {
	var app App
	var buf bytes.Buffer
	file, err := fs.Open(c)
	if err != nil {
		return &app
	}
	_, err = buf.ReadFrom(file)
	if err != nil {
		return &app
	}

	err = yaml.Unmarshal(buf.Bytes(), &app)
	if err != nil {
		return &app
	}

	ca := strings.Split(app.Cmd, " ")
	app.command = exec.Command(ca[0], ca[1:]...)
	app.updateVersion()
	return &app
}

func init() {
	fs = &afero.OsFs{}
}

func gitPull() error {
	var b bytes.Buffer

	pull := exec.Command("git", "pull", "-q")
	pull.Stderr = &b
	pull.Stdout = &b
	if err := pull.Run(); err != nil {
		log.Errorf("Git pull failed! %s", b.String())
		return err
	}
	return nil
}

func (a *App) updateVersion() error {
	if err := os.Chdir(a.Dir); err != nil {
		return err
	}

	var b bytes.Buffer
	rev := exec.Command("git", "rev-parse", "HEAD")
	rev.Stdout = &b
	rev.Stderr = &b
	if err := rev.Run(); err != nil {
		log.Errorf("Git rev-parse failed! %v", err)
		log.Error(b.String())
		return err
	}

	newRev := strings.TrimSpace(b.String())
	log.Debugf("git rev-parse: %s", newRev)
	if a.version != newRev {
		a.version = newRev
		a.restart = true
		log.Infof("%s version set to %s.", a.Name, a.version)
	}
	return nil
}

func (a *App) eql(b *App) bool {
	return a.Name == b.Name && a.Dir == b.Dir && a.Cmd == b.Cmd
}

func (a *App) update(t time.Time) error {
	log.Infof("Starting update of %s at %s", a.Name, t.String())
	a.restart = false

	if err := os.Chdir(a.Dir); err != nil {
		log.Errorf("Could not cd into %s: %v", a.Dir, err)
		return err
	}

	if err := gitPull(); err != nil {
		return err
	}

	a.updateVersion()
	if a.restart {
		log.Infof("Downloaded new version of %s, restarting it!", a.Name)

		var b bytes.Buffer
		a.command.Stdout = &b
		a.command.Stderr = &b

		err := a.command.Run()
		if err != nil {
			log.Errorf("Failed to run %q: %v", a.Cmd, err)
			log.Error(b.String())
			return err
		}
	}
	return nil
}
