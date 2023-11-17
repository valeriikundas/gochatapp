package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

func NewConfig(configName string) *Config {
	config := &Config{
		Viper: viper.New(),
	}

	if configName != "prod_config" {
		config.AddConfigPath(".")
		config.SetConfigName(configName)
		config.SetConfigType("yaml")

	}

	config.AutomaticEnv()

	err := config.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			if configName != "prod_config" {
				slog.Info("config file not found", "configName", configName, "err", err)
			}
		} else {
			panic(fmt.Errorf("fatal error: reading config file: %w", err))
		}
	}

	return config
}
