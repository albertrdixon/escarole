package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/ghodss/yaml"
	"golang.org/x/net/context"
)

func Read(file string) (*Config, error) {
	info, er := os.Stat(file)
	if er != nil {
		return conf, er
	}

	return read(file, info)
}

func ReadAndWatch(file string, ctx context.Context) (*Config, error) {
	c, er := Read(file)
	if er != nil {
		return c, er
	}

	go func() {
		logger.Debugf("Watching for config changes: config=%q", file)
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if er, ok := c.update(); er == nil && ok {
					logger.Infof("Config updated")
				}
			}
		}
	}()

	return c, nil
}

func (c *Config) update() (error, bool) {
	info, er := os.Stat(c.file)
	if er != nil {
		return er, false
	}

	if info.ModTime().Equal(c.modTime) {
		return nil, false
	}

	nc, er := read(c.file, info)
	if er == nil {
		c = nc
		return nil, true
	}
	return er, false
}

func read(file string, info os.FileInfo) (*Config, error) {
	logger.Debugf("Reading config from %q", file)
	content, er := ioutil.ReadFile(file)
	if er != nil {
		return nil, er
	}

	c := new(Config)
	if er := yaml.Unmarshal(content, c); er != nil {
		return conf, er
	}

	c.file = file
	c.modTime = info.ModTime()
	return c, nil
}
