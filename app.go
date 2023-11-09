package main

import (
	"flag"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var DB *gorm.DB
var validate *validator.Validate

func main() {
	shouldGenerateChats := flag.Bool("generateChats", false, "Should generate chats?")
	flag.Parse()

	_, err := os.Stat("uploads/")
	if os.IsNotExist(err) {
		os.MkdirAll("./uploads", 0744)
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
	log.Fatal(app.Listen("0.0.0.0:8080"))
}
