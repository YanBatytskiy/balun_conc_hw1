package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"envLocal"`
}

func MustLoad() *Config {
	const configPath = "config/yaml/local.yaml"

	_, err := os.Stat(configPath)
	if err != nil {
		log.Fatalf("error accessing config file: %s", err)
	}

	var cfg Config

	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg
}
