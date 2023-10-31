package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
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
	htmlEngine := html.New("templates/", ".html")

	app := fiber.New(fiber.Config{
		Views:       htmlEngine,
		ViewsLayout: "layouts/base",
		// Global custom error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResponse{
				Success: false,
				Message: err.Error(),
			})
		},
	})

	setUpRoutes(app)

	// app.Get("/login", func(c *fiber.Ctx) {
	// 	c.HTML(http.StatusOK, "lofiber.html", fiber.H{
	// 		"hello": "world",
	// 	})
	// })

	// app.Post("/login", func(c *fiber.Ctx) {
	// 	bytes, err := io.ReadAll(c.Request.Body)
	// 	if err != nil {
	// 		c.AbortWithError(http.StatusBadRequest, err)
	// 		return
	// 	}
	// 	data := string(bytes)
	// 	values, err := url.ParseQuery(data)
	// 	if err != nil {
	// 		c.AbortWithError(http.StatusBadRequest, err)
	// 	}
	// 	googleAuthResponse := GoogleAuthResponse{
	// 		credential:   values.Get("credential"),
	// 		g_csrf_token: values.Get("g_csrf_token"),
	// 	}
	// 	fmt.Printf("%v\n", googleAuthResponse)

	// 	c.IndentedJSON(http.StatusOK, fiber.H{
	// 		"login": "success",
	// 		"data": map[string]string{
	// 			"credential":   googleAuthResponse.credential,
	// 			"g_csrf_token": googleAuthResponse.g_csrf_token,
	// 		},
	// 	})
	// })

	return app
}

func setUpRoutes(app *fiber.App) {
	app.Get("/", HomeHandler)
	app.Get("/chats", ChatsViewHandler)
	app.Get("/chats/:chatID", ViewChat)

	app.Get("/users", GetUsersHandler)
	app.Post("/user", CreateUserHandler)
	app.Get("/api/chats", GetChatsHandler)
}
