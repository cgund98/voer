package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerPort string `default:"8080"`

	SqliteDBPath string `default:""`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("VOER", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
