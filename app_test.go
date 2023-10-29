package main

import (
	"math"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func TestAbs(t *testing.T) {
	t.Parallel()
	got := int(math.Abs(-1))
	if got != 1 {
		t.Errorf("Abs(-1) = %d; want 1", got)
	}
}

func TestFiber(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/users", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetChats(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/chats", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}
