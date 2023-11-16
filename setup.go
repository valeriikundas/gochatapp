package main

import (
	"flag"
	"fmt"
	"io"
	"os"

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

func connectDatabase(dbName string) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable TimeZone=Europe/Kiev", "0.0.0.0", 5432, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&User{}, &Chat{}, &Message{})
	if err != nil {
		panic(err)
	}
	return db
}

func createApp(db *gorm.DB, redisDB *redis.Storage) *fiber.App {
	setupLogger()

	htmlEngine := html.New("templates/", ".html")

	// TODO: will use django engine for new templates likely
	// djangoEngine := django.New("templates/django/", ".html")

	app := fiber.New(fiber.Config{
		AppName:     "GoChatApp",
		Views:       htmlEngine,
		ViewsLayout: "layouts/base",
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

		return c.Next()
	})

	if !isTesting() {
		app.Use(logger.New())
	}

	app.Static("/", "uploads/", fiber.Static{})

	setupRoutes(app)

	return app
}

func setupLogger() log.AllLogger {
	logger := log.DefaultLogger()

	logFile, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logWriter := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(logWriter)
	return logger
}

func isTesting() bool {
	return flag.Lookup("test.v") != nil
}
