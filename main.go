package main

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/albertrdixon/gearbox/logger"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("escarole", "Keeps your app leafy fresh!")
	project  = app.Arg("project", "github project. Format: Organization/Project, e.g. albertrdixon/escarole").Required().String()
	name     = app.Arg("name", "app name. If not given will use lowercase project name, e.g. Org/MyProject -> myproject").String()
	conf     = app.Flag("config", "path to command config").Short('C').Default("/escarole.yml").OverrideDefaultFromEnvar("CONFIG").ExistingFile()
	branch   = app.Flag("branch", "branch to use").Short('b').OverrideDefaultFromEnvar("BRANCH").String()
	interval = app.Flag("update-interval", "app update interval. Must be able to be parsed by time.ParseDuration").Short('u').Default("24h").OverrideDefaultFromEnvar("UPDATE_INTERVAL").Duration()
	uid      = app.Flag("uid", "app uid").Default("0").OverrideDefaultFromEnvar("APP_UID").Uint32()
	gid      = app.Flag("gid", "app gid").Default("0").OverrideDefaultFromEnvar("APP_GID").Uint32()
	env      = app.Flag("env", "app env vars").Short('e').PlaceHolder("key=value").StringMap()
	logLevel = app.Flag("log-level", "log level.").Short('l').PlaceHolder("{debug,info,warn,error,fatal}").Default("info").OverrideDefaultFromEnvar("LOG_LEVEL").Enum(logger.Levels...)

	git    string
	sha    string
	ref    string
	home   = "/src"
	stdout = []io.Writer{os.Stdout}
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	kingpin.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logger.Configure(*logLevel, "[escarole] ", os.Stdout)
	logger.Infof("Picking Escarole %v, so leafy!", version)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx, quit := context.WithCancel(context.Background())
	go func() {
		select {
		case s := <-sig:
			logger.Infof("Got signal %v, terminating", s)
			quit()
			time.Sleep(100 * time.Millisecond)
			os.Exit(0)
		}
	}()

	if er := setup(ctx); er != nil {
		quit()
		logger.Fatalf("Setup failed: %v", er)
	}

	app, er := prepareApp(ctx)
	if er != nil {
		quit()
		logger.Fatalf(er.Error())
	}
	go run(app, ctx, quit)

	<-ctx.Done()
}

func setup(c context.Context) error {
	logger.Infof("Setting HOME to %s", home)
	if er := os.Setenv("HOME", home); er != nil {
		logger.Warnf(er.Error())
	}

	logger.Infof("Caching git binary location")
	g, er := exec.LookPath("git")
	if er != nil {
		logger.Fatalf("git not found in path: %v", er)
	}
	git = g

	if er := clone(c); er != nil {
		return er
	}

	s, er := getSHA()
	if er != nil {
		logger.Fatalf("Failed to get sha: %v", er)
	}
	r, er := getRef()
	if er != nil {
		logger.Fatalf("Failed to get ref: %v", er)
	}

	sha = s
	ref = r
	return nil
}
