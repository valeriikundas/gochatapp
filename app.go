package main

import (
	"flag"
	"fmt"
	"log/slog"

	"os"

	"github.com/gofiber/fiber/v2/log"
)

func main() {
	// FIXME: write test for `main` function
	// FIXME: make it enum
	envOptions := []string{"dev", "docker", "test", "prod"}
	env := flag.String("config", "prod", fmt.Sprintf("What config to use. Options are %v", envOptions))
	shouldGenerateChats := flag.Bool("generateChats", false, "Should generate chats?")
	flag.Parse()

	slog.Info("initial read env", "env", env)
	if env == nil {
		*env = "prod"
	}
	if !contains(envOptions, env) {
		panic(fmt.Errorf("unknown env: %v", env))
	}

	config := NewConfig(fmt.Sprintf("%s_config", *env))
	if config == nil {
		panic(fmt.Errorf("failed reading config for env=%s", *env))
	}

	_, err := os.Stat("uploads/")
	if os.IsNotExist(err) {
		err = os.MkdirAll("./uploads", 0744)
		if err != nil {
			panic(err)
		}
	}

	postgresDB := getPostgres(config)

	if *shouldGenerateChats {
		err := generateRandomChats(nil, postgresDB)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	redisDB := getRedis(config)

	app := createApp(postgresDB, redisDB)

	appUrl := getAppURL(config)
	log.Fatal(app.Listen(appUrl))

}

func contains(envOptions []string, env *string) bool {
	for _, v := range envOptions {
		if v == *env {
			return true
		}
	}
	return false
}

func getAppURL(config *Config) string {
	return "0.0.0.0:3000"
}
