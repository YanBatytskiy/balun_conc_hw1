package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Network *NetworkConfig `yaml:"network"`
	Logger  *LoggingConfig `yaml:"logging"`

	EngineType string `yaml:"engine_type" env-default:"in_memory"`
}

type NetworkConfig struct {
	Address        string        `yaml:"engine_address"  env-default:"127.0.0.1:3323"`
	MaxConnections int           `yaml:"max_connections" env-default:"100" validate:"gt=0"`
	MaxMessageSize int           `yaml:"max_message_size" env-default:"4096" validate:"gt=0"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"    env-default:"5m" validate:"gt=0"`
	BufferSize     int           `yaml:"buffer_size"     env-default:"4096" validate:"gt=0"`
	TypeConn       string        `yaml:"type" env-default:"tcp"`
}

type LoggingConfig struct {
	Level             string `yaml:"level"               env-default:"info"`
	LogFile           string `yaml:"output"              env-default:"/log/output.log"`
	LoggerStoragePath string `yaml:"logger_storage_path" env-default:"./spyder/storage/log_storage.db"`
	LoggerAddress     string `yaml:"logger_address"      env-default:"127.0.0.1:8082"`
	LoggerUser        string `yaml:"user"                env-default:"myuser"`
	LoggerPass        string `yaml:"password"            env-default:"mypassword"`
}

func NewConfig() (*Config, error) {
	const op = "config.NewConfig"

	configPath := os.Getenv("CONFIG_PATH")

	var cfg = Config{
    Network: &NetworkConfig{
        Address:        "127.0.0.1:3323",
        MaxConnections: 100,
        MaxMessageSize: 4096,
        IdleTimeout:    5 * time.Minute,
        BufferSize:     4096,
        TypeConn:       "tcp",
    },
    EngineType: "in_memory",
    Logger: &LoggingConfig{
        Level:             "info",
        LogFile:           "/log/output.log",
        LoggerStoragePath: "./spyder/storage/log_storage.db",
        LoggerAddress:     "127.0.0.1:8082",
        LoggerUser:        "myuser",
        LoggerPass:        "mypassword",
    },
}


	if configPath != "" {
		_, err := os.Stat(configPath)
		if err != nil {
			return nil, fmt.Errorf("%s: wrong path: error accessing config file", op)
		}

		err = cleanenv.ReadConfig(configPath, &cfg)
		if err != nil {
			return nil, fmt.Errorf("%s: error reading config file", op)
		}
	} else {
		err := cleanenv.ReadEnv(&cfg)
		if err != nil {
			return nil, fmt.Errorf("%s: error reading config from env variables", op)
		}
	}

	err := validator.New().Struct(&cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: config validation error: %w", op, err)
	}
	
	return &cfg, nil
}
