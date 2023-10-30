package main

import (
	"flag"
	"log"

	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	shouldGenerateChats := flag.Bool("generateChats", false, "Should generate chats?")
	flag.Parse()

	db = connectDatabase("chatapp")

	if *shouldGenerateChats {
		err := generateRandomChats(nil, db)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	app := createApp(db)
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
