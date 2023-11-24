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
		CPUs        []uint16
		Polar       PolarConfig
		Parallelism uint16 `yaml:"-"`
		Debug       bool   `yaml:"-"`
		Key         string
		Storage     StorageConfig
	}
	PolarConfig struct {
		Host string
		Port uint16
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
