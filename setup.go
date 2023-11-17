package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
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

func connectDatabase(postgresHost string, postgresPort int, postgresUser string, postgresPassword string, postgresDatabase string) *gorm.DB {
	dsn := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(postgresUser, postgresPassword),
		Host:     fmt.Sprintf("%s:%d", postgresHost, postgresPort),
		Path:     postgresDatabase,
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("error connect to database: %w, dsn=%s", err, dsn.String()))
	}

	return db
}

func createApp(pgDB *gorm.DB, redisDB *redis.Storage) *fiber.App {
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
