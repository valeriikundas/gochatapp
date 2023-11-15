package main

import (
	"flag"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"gorm.io/gorm"
)

var DB *gorm.DB
var validate *validator.Validate
var redisStorage = redis.New(redis.Config{
	Host:     "0.0.0.0",
	Port:     6379,
	Username: "valeriikundas",
})
var store = session.New(session.Config{
	Storage: redisStorage,
})

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

	DB = connectDatabase("chatapp")

	if *shouldGenerateChats {
		err := generateRandomChats(nil, DB)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	app := createApp(DB)
	log.Fatal(app.Listen("0.0.0.0:3000"))
}
