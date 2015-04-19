package main

import (
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	debug    = kingpin.Flag("debug", "Enable debug output.").Short('d').Bool()
	interval = kingpin.Flag("interval", "Set update interval. Must be parseable by time.ParseTime (e.g. 20m, 2h, etc.).").Short('i').Default("30m").OverrideDefaultFromEnvar("UPDATE_INTERVAL").Duration()
	conf     = kingpin.Flag("conf", "Escarole config. Must be a real file.").Short('C').Default("/etc/escarole.yaml").OverrideDefaultFromEnvar("UPDATE_CONF").ExistingFile()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	log.WithField("config", *conf).Info("Picking Escarole! So leafy!")

	app := newApp(*conf)
	tick := time.NewTicker(*interval)
	for {
		select {
		case <-sig:
			os.Exit(0)
		case t := <-tick.C:
			for i := 0; i < 3; i++ {
				err := app.update(t)
				if err == nil {
					break
				}
				time.Sleep(20 * time.Second)
			}
		}
	}
}
