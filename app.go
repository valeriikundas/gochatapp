package main

import (
	"flag"
	"log"
	"os"

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

// TODO: add photo storage: 2 options: postgresql, s3-like storage
// implement 2 options with interchangibility option
// maybe use cache for storing lately accessed images

// TODO: use tdd, coverage to >90%
// TODO: add auth https://medium.com/@abhinavv.singh/a-comprehensive-guide-to-authentication-and-authorization-in-go-golang-6f783b4cea18
// TODO: use JWT tokens
// TODO: in general writing a chat app
// TODO: write a signup in tdd fashion
// TODO: add cache (e.g. redis)
// TODO: add pubsub/message queue of some kind
// TODO: auth with permissions roles (user, admin, chat admin)
// TODO: setup swagger https://docs.gofiber.io/contrib/swagger_v1.x.x/swagger/
// TODO: try https://github.com/go-gorm/gen
// TODO: use protobufs
// TODO: setup monkey testing
// TODO: try BDD
// TODO: deploy to aws
// TODO: setup docker-compose
// TODO: setup ci/cd
// TODO: use all features of gorm
// TODO: use all features of fiber
// TODO: setup automatic backup for database and images
// TODO: use db hooks for something
// TODO: add chatgpt integration as bot
// TODO: learn to use os, os/exec, io, bytes libs
// TODO: learn to use tags
// TODO: use goroutines somewhere
// TODO: play around and debug internaly of fiber, gorm, docker etc
// TODO: use github copilot free trial
// TODO: use factories for tests. is it useful in golang? https://github.com/bluele/factory-go
// TODO: setting up db for tests - setup, teardown - in other projects
// TODO: setup database migrations
// TODO: try https://github.com/sqlc-dev/sqlc and maybe benchmark
// TODO: write raw SQL query
// TODO: review templates e.g https://github.com/create-go-app/fiber-go-template/tree/master and others
// TODO: try supabase
// TODO: add payments
// TODO: implement rate limiting and maybe some more features from distributed applications :)
// TODO: try planetscale database
// TODO: structure logging https://pkg.go.dev/golang.org/x/exp/slog
// TODO: setup sentry or some other monitoring
// TODO: try rpc, grpc, webrtc
// TODO: setup linter for function length and code complexity
// TODO: setup load testing https://github.com/tsenart/vegeta
// TODO: add api versioning
