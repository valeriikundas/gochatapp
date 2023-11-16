package main

import (
	"flag"
	"log"
	"os"

	"github.com/gofiber/storage/redis/v3"
)

func main() {
	shouldGenerateChats := flag.Bool("generateChats", false, "Should generate chats?")
	flag.Parse()

	_, err := os.Stat("uploads/")
	if os.IsNotExist(err) {
		err = os.MkdirAll("./uploads", 0744)
		if err != nil {
			panic(err)
		}
	}

	pgDB := connectDatabase("chatapp")

	if *shouldGenerateChats {
		err := generateRandomChats(nil, pgDB)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	redisDB := redis.New(redis.Config{
		Host:     "0.0.0.0",
		Port:     6379,
		Username: "valeriikundas",
		Database: 0,
	})

	app := createApp(pgDB, redisDB)
	log.Fatal(app.Listen("0.0.0.0:3000"))
}
