package main

import (
	"flag"
	"log"

	"github.com/go-playground/validator/v10"
	fiberlog "github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

var DB *gorm.DB
var validate *validator.Validate
var logger fiberlog.AllLogger

func main() {
	shouldGenerateChats := flag.Bool("generateChats", false, "Should generate chats?")
	flag.Parse()

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

// TODO: add auth https://medium.com/@abhinavv.singh/a-comprehensive-guide-to-authentication-and-authorization-in-go-golang-6f783b4cea18
// TODO: generate random chats
// TODO: first test
// TODO: in general writing a chat app
// TODO: write a signup in tdd fashion
// TODO: add cache
// TODO: add pubsub
// TODO: add photo storage
// TODO: auth with permissions roles (user, admin, chat admin)
// TODO: setup https://docs.gofiber.io/contrib/swagger_v1.x.x/swagger/
