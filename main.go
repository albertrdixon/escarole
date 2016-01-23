package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/albertrdixon/escarole/config"
	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/gearbox/process"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("escarole", "Keeps your app leafy fresh")
	interval = app.Flag("update-interval", "Set update interval. Must be parseable by time.ParseTime (e.g. 20m, 2h, etc.).").Short('u').Default("6h").OverrideDefaultFromEnvar("UPDATE_INTERVAL").Duration()
	conf     = app.Flag("config", "Escarole config. Must be a real file.").Short('C').Default("/etc/escarole.yaml").OverrideDefaultFromEnvar("UPDATE_CONF").ExistingFile()
	logLevel = app.Flag("log-level", "log level. One of: fatal, error, warn, info, debug").Short('l').Default("info").OverrideDefaultFromEnvar("LOG_LEVEL").Enum(logger.Levels...)
)

func execute(c *config.Config, ctx context.Context) {
	app := process.New(c.Name, c.Command, os.Stdout, ctx)

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	kingpin.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	logger.Configure(logLevel, "[escarole] ", os.Stdout)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	logger.Infof("Picking Escarole %v, so leafy! Config: %v", version, *conf)
	c, q := context.WithCancel(context.Background())
	conf, er := config.ReadAndWatch(*conf, c)

	app, err := newApp(*conf)
	if err != nil {
		log.Warnf("Could not load config: %v", err)
	}
	log.Infof("Will be growing %v", app)

	tick := time.NewTicker(*interval)
	log.Infof("Started timer. interval: %v", *interval)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-tick.C:
				go func() {
					for i := 0; i < 3; i++ {
						err := app.update(t)
						if err == nil {
							break
						}
						time.Sleep(20 * time.Second)
					}
				}()
			}
		}
	}()

	<-sig
	log.Info("Shutting down...")
	done <- struct{}{}
	os.Exit(0)
}
