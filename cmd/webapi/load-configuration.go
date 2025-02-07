package main

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// WebAPIConfiguration describes the web API configuration.
type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:/conf/config.yml"`
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	Debug bool
	DB    struct {
		Filename string `conf:"default:/tmp/decaf.db"`
	}
}

// loadConfiguration creates a WebAPIConfiguration from flags, environment variables and a configuration file.
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// Parse environment variables and CLI flags.
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err2 := conf.Usage("CFG", &cfg)
			if err2 != nil {
				return cfg, err2
			}
			// Use logger instead of fmt.Println
			logrus.Info(usage)
			return cfg, conf.ErrHelpWanted
		}
		return cfg, err
	}

	// Override with YAML config file if it exists.
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, err
	} else if err == nil {
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, err
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			return cfg, err
		}
		_ = fp.Close()
	}

	return cfg, nil
}
