package config

import (
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

type (
	IglooConfig struct {
		ID          string        `yaml:"id"`
		CPUs        []uint16      `yaml:"cpus"`
		Polar       PolarConfig   `yaml:"polar"`
		Parallelism uint16        `yaml:"-"`
		Debug       bool          `yaml:"-"`
		Storage     StorageConfig `yaml:"storage"`
	}
	PolarConfig struct {
		Host       string `yaml:"host"`
		Port       uint16 `yaml:"port"`
		SecretHash string `yaml:"secretHash"`
	}
	StorageConfig struct {
		Submissions string `yaml:"submissions"`
		Problems    string `yaml:"problems"`
	}
)

var Config = new(IglooConfig)

func init() {
	confPath := strings.TrimSpace(os.Getenv("IGLOO_CONFIG_PATH"))
	if confPath == "" {
		confPath = "igloo.yml"
	}
	b, e := os.ReadFile(confPath)
	logger.Panic(e, "failed to read config")
	logger.Panic(yaml.Unmarshal(b, Config), "failed to parse config")
	Config.Debug = os.Getenv("ENV") == "dev"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if Config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
