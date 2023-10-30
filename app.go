package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	*gorm.Model

	Name  string `gorm:"uniqueIndex" validate:"required"`
	Email string `gorm:"uniqueIndex" validate:"required"`

	Chats []Chat `gorm:"many2many:chat_members"`
}

type Chat struct {
	*gorm.Model

	Name    string `gorm:"uniqueIndex" validate:"required"`
	Members []User `gorm:"many2many:chat_members"`
}

// type GoogleAuthResponse struct {
// 	credential   string
// 	g_csrf_token string
// }

type GlobalErrorHandlerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type FieldError struct {
	Field, Tag, Param string
}

func main() {
	shouldGenerateChats := flag.Bool("generateChats", false, "generate chats?")
	flag.Parse()

	db := connectDatabase("chatapp")

	if *shouldGenerateChats {
		generateRandomChats(nil, db)
		return
	}

	app := createApp(db)
	log.Fatal(app.Listen("0.0.0.0:8080"))
}

func connectDatabase(dbName string) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable TimeZone=Europe/Kiev", "0.0.0.0", 5432, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&User{}, &Chat{})
	if err != nil {
		panic(err)
	}
	return db
}

func createApp(db *gorm.DB) *fiber.App {
	htmlEngine := html.New("templates/", ".html")

	app := fiber.New(fiber.Config{
		Views: htmlEngine,
		// Global custom error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResponse{
				Success: false,
				Message: err.Error(),
			})
		},
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("home", fiber.Map{
			"a": "b",
		})
	})

	// now: connected database, user table
	// next:
	// - setup testing
	// - login signup methods

	app.Get("/users", func(c *fiber.Ctx) error {
		var users []User
		tx := db.Find(&users)
		log.Printf("%v\n", users)
		if tx.Error != nil {
			return tx.Error
		}
		return c.Render("users", fiber.Map{
			"Users": users,
		})
	})

	app.Post("/user", func(c *fiber.Ctx) error {
		var user User
		err := c.BodyParser(&user)
		if err != nil {
			return err
		}
		validate := validator.New()
		err = validate.Struct(user)
		if err != nil {
			var errors []FieldError
			for _, err := range err.(validator.ValidationErrors) {
				el := FieldError{
					Field: err.Field(),
					Tag:   err.Tag(),
					Param: err.Param(),
				}
				errors = append(errors, el)
			}
			return c.Status(fiber.StatusBadRequest).JSON(errors)
		}
		tx := db.Create(&user)
		if tx.Error != nil {
			return tx.Error
		}
		return c.JSON(user)
	})

	app.Get("/api/chats", func(c *fiber.Ctx) error {
		var chats []Chat
		tx := db.Model(&Chat{}).Preload("Members").Find(&chats)
		if tx.Error != nil {
			return tx.Error
		}
		bytes, err := json.MarshalIndent(fiber.Map{"chats": chats}, "", "  ")
		if err != nil {
			return err
		}
		return c.SendString(string(bytes))

	})

	app.Get("/chats", func(c *fiber.Ctx) error {
		var chats []Chat
		tx := db.Find(&chats)
		if tx.Error != nil {
			return tx.Error
		}
		return c.Render("chats", fiber.Map{
			"chats": chats,
		})
	})

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

// TODO: add auth https://medium.com/@abhinavv.singh/a-comprehensive-guide-to-authentication-and-authorization-in-go-golang-6f783b4cea18
// TODO: generate random chats
// TODO: first test
// TODO: in general writing a chat app
// TODO: write a signup in tdd fashion
// TODO: add cache
// TODO: add pubsub
// TODO: add photo storage
