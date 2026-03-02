package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	Services struct {
		MoodTracker struct {
			URL     string        `yaml:"url"`
			Timeout time.Duration `yaml:"timeout"`
		} `yaml:"mood_tracker"`
	} `yaml:"services"`

	Limiter struct {
		RPS     float64 `yaml:"rps"`
		Burst   int     `yaml:"burst"`
		Enabled bool    `yaml:"enabled"`
	} `yaml:"limiter"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
