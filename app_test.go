package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	db, dropCmd := prepareTestDb(t)
	defer dropCmd.Run()
	app := createApp(db)

	generateRandomChats(t, db)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/chats", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	var data GetChatsResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 100, len(data.Chats))
}

func TestSendMessage(t *testing.T) {
	db1, dropCmd := prepareTestDb(t)
	db = db1
	defer dropCmd.Run()

	app := createApp(db)

	// TODO: mock database during testing
	// app := fiber.New(fiber.Config{})
	// app.Add(fiber.MethodPost, "/api/chat/:chatID", SendMessage)

	user, err := addRandomUser(db)
	utils.AssertEqual(t, nil, err)

	chat, err := addRandomChat(db)
	utils.AssertEqual(t, nil, err)

	buf := new(bytes.Buffer)
	messageContent := "hello"
	data := SendMessageRequest{
		FromID:  user.ID,
		ChatID:  chat.ID,
		Content: messageContent,
	}
	err = json.NewEncoder(buf).Encode(data)
	utils.AssertEqual(t, nil, err)

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/api/chat/%d", chat.ID), buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)

	bytes2, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	var sendMessageResponse struct {
		ID uint
	}
	err = json.Unmarshal(bytes2, &sendMessageResponse)
	utils.AssertEqual(t, nil, err)

	var message Message
	tx := db.Find(&message)
	utils.AssertEqual(t, nil, tx.Error)
	utils.AssertEqual(t, message.ID, sendMessageResponse.ID)
	utils.AssertEqual(t, messageContent, message.Content)
	utils.AssertEqual(t, user.ID, message.FromID)
	utils.AssertEqual(t, chat.ID, message.ChatID)
}
