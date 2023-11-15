package main

import (
	"encoding/json"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func IndentJSONResponseMiddleware(c *fiber.Ctx) error {
	err := c.Next()
	if err != nil {
		return err
	}

	contentType := c.Response().Header.ContentType()
	if string(contentType) != fiber.MIMEApplicationJSON {
		return nil
	}

	responseBody := c.Response().Body()
	var data any
	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		return err
	}

	contentType, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	c.Response().SetBody(contentType)
	return nil
}

func AssertWebSocketUpgradeMiddleware(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}
