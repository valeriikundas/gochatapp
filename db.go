package main

import (
	"github.com/gofiber/storage/redis/v3"
	"gorm.io/gorm"
)

func getPostgres(config *Config) *gorm.DB {
	postgresHost := config.GetString("postgres_host")
	postgresPort := config.GetInt("postgres_port")
	postgresDBName := config.GetString("postgres_dbname")

	postgresDB := connectDatabase(postgresHost, postgresPort, postgresDBName)

	// TODO: get a list of tables from somewhere
	err := postgresDB.AutoMigrate(&User{}, &Chat{}, &Message{})
	if err != nil {
		panic(err)
	}

	return postgresDB
}

func getRedis(config *Config) *redis.Storage {
	redisHost := config.GetString("redis_host")
	redisPort := config.GetInt("redis_port")
	redisUsername := config.GetString("redis_username")
	redisDatabase := config.GetInt("redis_database")

	redisDB := redis.New(redis.Config{
		Host:     redisHost,
		Port:     redisPort,
		Username: redisUsername,
		Database: redisDatabase,
	})

	return redisDB
}
