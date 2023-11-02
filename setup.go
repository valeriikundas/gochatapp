package main

import (
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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

func createApp(db *gorm.DB) *fiber.App {
	logger = setupLogger()

	validate = validator.New()

	htmlEngine := html.New("templates/", ".html")

	app := fiber.New(fiber.Config{
		Views:       htmlEngine,
		ViewsLayout: "layouts/base",
		// Global custom error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Errorf("global error = %v\n", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResponse{
				Success: false,
				Message: err.Error(),
			})
		},
	})

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
