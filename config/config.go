package config

import (
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"igloo/logger"
	"os"
)

type (
	IglooConfig struct {
		ID          string
		Debug       bool `yaml:"-"`
		Key         string
		Parallelism int16
		Storage     *StorageConfig
		RabbitMQ    *RabbitMQConfig
	}
	RabbitMQConfig struct {
		Username   string
		Password   string
		Host       string
		Port       uint16
		StreamPort uint16 `yaml:"streamPort"`
		VHost      string
	}
	StorageConfig struct {
		Submissions string
		Problems    string
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
