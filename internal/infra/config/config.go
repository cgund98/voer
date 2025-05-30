package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Client-side value for the server's gRPC endpoint
	GrpcEndpoint string `default:"localhost:8000"`

	// Server-side value for the port to listen on
	GrpcPort     int `default:"8000"`
	FrontendPort int `default:"8080"`

	// Path to the sqlite3 database file
	SqliteDBPath string `default:""`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("VOER", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
