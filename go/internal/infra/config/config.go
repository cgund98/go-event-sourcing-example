package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GrpcPort int `default:"8081"`
	HttpPort int `default:"8080"`

	PostgresHost     string `default:"localhost"`
	PostgresPort     int    `default:"5432"`
	PostgresUser     string `default:"postgres"`
	PostgresPassword string `default:"postgres"`
	PostgresDB       string `default:"orders"`

	EventsTable string `default:"events"`
	EventsTopic string `default:"events"`

	KafkaHost string `default:"localhost"`
	KafkaPort int    `default:"9092"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("ORDER_SVC", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
