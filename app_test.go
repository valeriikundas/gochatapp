package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
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
	app, db, teardownTest := setupTest(t)
	defer teardownTest()

	err := generateRandomChats(t, db)
	utils.AssertEqual(t, nil, err)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/chats", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	var data GetChatsResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 100, len(data.Chats))
}

func TestSendMessage(t *testing.T) {
	app, db, teardownTest := setupTest(t)
	defer teardownTest()

	// TODO: mock database during testing
	// app := fiber.New(fiber.Config{})
	// app.Add(fiber.MethodPost, "/api/chat/:chatID", SendMessage)

	user, err := addRandomUser(db, false)
	utils.AssertEqual(t, nil, err)

	chat, err := addRandomChatWithNoUsers(db)
	utils.AssertEqual(t, nil, err)

	err = db.Model(&chat).Association("Members").Append(&user)
	utils.AssertEqual(t, nil, err, "Chat add Member")
	err = db.Save(&chat).Error
	utils.AssertEqual(t, nil, err, "Chat save after add Member")

	// login user
	sessionCookie := getLoggedInUserSessionCookie(t, app, *user)

	messageContent := "hello"
	data := SendMessageRequest{
		UserEmail: user.Email,
		Content:   messageContent,
	}
	marshalled, err := json.Marshal(data)
	utils.AssertEqual(t, nil, err)
	buf1 := bytes.NewReader(marshalled)

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/api/chats/%d", chat.ID), buf1)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  sessionCookie.Name,
		Value: sessionCookie.Value,
	})
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	bytes2, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	var sendMessageResponse struct {
		ID uint
	}
	err = json.Unmarshal(bytes2, &sendMessageResponse)
	utils.AssertEqual(t, nil, err, fmt.Sprintf("%+v", sendMessageResponse))

	var message Message
	tx := db.Find(&message)
	utils.AssertEqual(t, nil, tx.Error)
	utils.AssertEqual(t, message.ID, sendMessageResponse.ID)
	utils.AssertEqual(t, messageContent, message.Content)
	utils.AssertEqual(t, user.ID, message.FromID)
	utils.AssertEqual(t, chat.ID, message.ChatID)
}
