package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	config := NewConfig("dev_config")

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

func getAppURL(config *Config) string {
	appHost := config.GetString("app_host")
	appPort := config.GetInt("app_port")

	appUrl := fmt.Sprintf("%s:%d", appHost, appPort)
	return appUrl
}
