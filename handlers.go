package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func ChatsViewHandler(c *fiber.Ctx) error {
	var chats []Chat
	tx := db.Find(&chats)
	if tx.Error != nil {
		return tx.Error
	}
	return c.Render("chats", fiber.Map{
		"chats": chats,
	})
}

func HomeHandler(c *fiber.Ctx) error {
	return c.Render("home", fiber.Map{
		"a": "b",
	})
}

func GetUsersHandler(c *fiber.Ctx) error {
	var users []User
	tx := db.Find(&users)
	log.Printf("%v\n", users)
	if tx.Error != nil {
		return tx.Error
	}
	return c.Render("users", fiber.Map{
		"Users": users,
	})
}

type FieldError struct {
	Field, Tag, Param string
}

func CreateUserHandler(c *fiber.Ctx) error {
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
}

type GetChatsResponse struct {
	Chats []Chat
}

func GetChatsHandler(c *fiber.Ctx) error {
	var chats []Chat
	tx := db.Model(&Chat{}).Preload("Members").Find(&chats)
	if tx.Error != nil {
		return tx.Error
	}

	data := GetChatsResponse{
		Chats: chats,
	}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(fiber.Map{"chats": chats}, "", "  ")
	if err != nil {
		return err
	}

	return c.SendString(string(bytes))
}
