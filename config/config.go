package config

import (
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"os"
)

type (
	IglooConfig struct {
		ID          string
		Debug       bool `yaml:"-"`
		Key         string
		Parallelism int16
		Storage     StorageConfig
		RabbitMQ    RabbitMQConfig
		// TODO: make caching optional
		Dragonfly Address
	}
	RabbitMQConfig struct {
		Username   string
		Password   string
		Address    `yaml:",inline"`
		StreamPort uint16 `yaml:"streamPort"`
		VHost      string
	}
	StorageConfig struct {
		Submissions string
		Problems    string
	}
	Address struct {
		Host string
		Port uint16
	}
)

var Config = new(IglooConfig)

func init() {
	b, e := os.ReadFile("igloo.yml")
	logger.Panic(e, "failed to read config")
	logger.Panic(yaml.Unmarshal(b, Config), "failed to parse config")
	Config.Debug = os.Getenv("ENV") == "dev"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if Config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
