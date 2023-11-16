package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func handleValidationError(c *fiber.Ctx, err error) error {
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
