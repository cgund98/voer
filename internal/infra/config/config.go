package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Client-side value for the server's gRPC endpoint
	GrpcEndpoint string `default:"localhost:8000"`
	GrpcTls      bool   `default:"true"`

	// Server-side value for the port to listen on
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
