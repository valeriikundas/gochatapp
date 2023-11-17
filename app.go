package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	envOptions := []string{"dev", "docker", "test"}
	env := flag.String("config", "dev", fmt.Sprintf("What config to use. Options are %v", envOptions))
	if env == nil {
		*env = "dev"
	}
	if !contains(envOptions, env) {
		panic(fmt.Errorf("unknown env: %v", env))
	}

	config := NewConfig(fmt.Sprintf("%s_config", *env))
	if config == nil {
		panic(fmt.Errorf("failed reading config for env=%s", *env))
	}

	shouldGenerateChats := flag.Bool("generateChats", false, "Should generate chats?")
	flag.Parse()

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
	appHost := config.GetString("app_host")
	appPort := config.GetInt("app_port")

	appUrl := fmt.Sprintf("%s:%d", appHost, appPort)
	return appUrl
}
