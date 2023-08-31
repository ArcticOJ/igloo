package config

import (
	"gopkg.in/yaml.v2"
	"igloo/igloo/logger"
	"os"
)

type (
	IglooConfig struct {
		Address     `yaml:",inline"`
		ID          string
		Key         string
		Parallelism int16
		Storage     *StorageConfig
		RabbitMQ    *RabbitMQConfig
	}
	RabbitMQConfig struct {
		Username string
		Password string
		Address  `yaml:",inline"`
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
}
