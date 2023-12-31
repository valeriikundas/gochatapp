package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/storage/redis/v3"
	"gorm.io/gorm"
)

// FIXME: pass connection url as argument to both postgres&redis functions

func getPostgres(config *Config) *gorm.DB {
	var dsn string

	if config.ConfigFileUsed() == "prod_config" {
		dsn = os.Getenv("DATABASE_URL")
		slog.Debug("postgres: using prod_config", "dsn", dsn, "config", config.ConfigFileUsed())
		slog.Debug("postgres: using prod_config", "dsn", dsn, "config", config.ConfigFileUsed())
	} else {
		dsn = config.GetString("DATABASE_URL")
		slog.Debug("postgres: using non-prod config", "config", config.ConfigFileUsed(), "dsn", dsn)
		slog.Debug("postgres: using non-prod config", "config", config.ConfigFileUsed(), "dsn", dsn)
	}
	slog.Debug("connect postgres", "url", dsn, "config", config.ConfigFileUsed())
	postgresDB := connectDatabase(dsn)

	// TODO: get a list of tables from somewhere
	err := postgresDB.AutoMigrate(&User{}, &Chat{}, &Message{})
	if err != nil {
		panic(err)
	}

	return postgresDB
}

func getRedis(config *Config) *redis.Storage {
	var redisURL string

	if config.ConfigFileUsed() == "prod_config" {
		redisURL = os.Getenv("REDIS_URL")
		slog.Debug("redis: using prod_config", "redisURL", redisURL)
		slog.Debug("redis: using prod_config", "redisURL", redisURL, "config", config.ConfigFileUsed())
	} else {
		redisURL = config.GetString("REDIS_URL")
		slog.Debug("redis: using non-prod config", "redisURL", redisURL)
		slog.Debug("redis: using non-prod config", "redisURL", redisURL, "config", config.ConfigFileUsed())

	}

	slog.Debug("connect", "redisURL", redisURL)
	redisDB := redis.New(redis.Config{
		URL: redisURL,
	})
	return redisDB
}
