package config

import (
	"fmt"
	"go.yaml.in/yaml/v4"
	"os"
	"time"
)

type BackendCfg struct {
	Url    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

type TimeoutsCfg struct {
	Dial        time.Duration `yaml:"dial"`
	Connect     time.Duration `yaml:"connect"`
	Shutdown    time.Duration `yaml:"shutdown"`
	HealthCheck time.Duration `yaml:"health_check"`
}

type Config struct {
	Port      int          `yaml:"port"`
	Algorithm string       `yaml:"algorithm"`
	Backends  []BackendCfg `yaml:"backends"`
	Timeouts  TimeoutsCfg  `yaml:"timeouts"`
}

func Load(filePath string) Config {
	var conf Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(data, &conf); err != nil {
		panic(err)
	}
	if err = conf.validate(); err != nil {
		panic(err)
	}
	return conf
}

func (c *Config) validate() error {
	for i, b := range c.Backends {
		if b.Url == "" {
			return fmt.Errorf("backend %d: url required", i)
		}
	}
	return nil
}
