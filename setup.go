package main

import (
	"embed"
	"flag"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GlobalErrorHandlerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func connectDatabase(DSN string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("error connect to database: %w, dsn=%s", err, DSN))
	}

	return db
}

//go:embed templates/*
var templatesFS embed.FS

func createApp(pgDB *gorm.DB, redisDB *redis.Storage) *fiber.App {
	htmlEngine := html.NewFileSystem(http.FS(templatesFS), ".html")

	app := fiber.New(fiber.Config{
		AppName:     "GoChatApp",
		Views:       htmlEngine,
		ViewsLayout: "templates/layouts/base",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Errorf("global error = %v\n", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResponse{
				Success: false,
				Message: err.Error(),
			})
		},
	})

	app.Use(IndentJSONResponseMiddleware)
	app.Use("/ws", AssertWebSocketUpgradeMiddleware)

	app.Use(func(c *fiber.Ctx) error {
		validate := validator.New()
		c.Locals("validate", validate)

		// TODO: rename to `sessionStore` soon
		var store = session.New(session.Config{
			Storage: redisDB,
		})
		c.Locals("store", store)

		c.Locals("db", pgDB)

		return c.Next()
	})

	if !isTesting() {
		app.Use(logger.New())
	}

	app.Static("/", "uploads/", fiber.Static{})

	setupRoutes(app)

	return app
}

func isTesting() bool {
	return flag.Lookup("test.v") != nil
}
