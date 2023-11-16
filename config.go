package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

func NewConfig(configName string) *Config {
	config := &Config{
		Viper: viper.New(),
	}

	config.AddConfigPath(".")
	config.SetConfigName(configName)
	config.SetConfigType("yaml")

	config.AutomaticEnv()

	err := config.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			log.Debugf("fatal error config file not found: %w", err)
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	return config
}
