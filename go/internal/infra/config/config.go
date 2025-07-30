package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GrpcPort int `default:"9092"`
	HttpPort int `default:"8080"`

	PostgresHost     string `default:"localhost"`
	PostgresPort     int    `default:"5432"`
	PostgresUser     string `default:"postgres"`
	PostgresPassword string `default:"postgres"`
	PostgresDB       string `default:"orders"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("ORDER_SVC", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
