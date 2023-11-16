package main

import (
	"flag"
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
	log.Fatal(app.Listen("0.0.0.0:3000"))
}
