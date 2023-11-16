package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
)

type Configuration struct {
	*viper.Viper
}

func NewConfig(configName string) *Configuration {
	configuration := &Configuration{
		Viper: viper.New(),
	}

	configuration.AddConfigPath(".")
	configuration.SetConfigName(configName)
	configuration.SetConfigType("yaml")

	configuration.AutomaticEnv()

	err := configuration.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			log.Debugf("fatal error config file not found: %w", err)
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	return configuration
}
