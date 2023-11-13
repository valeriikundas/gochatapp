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
	"gorm.io/gorm"
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
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)
	app := createApp(DB)

	generateRandomChats(t, DB)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/chats", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	var data GetChatsResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 100, len(data.Chats))
}

func TestSendMessage(t *testing.T) {
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)

	app := createApp(DB)

	// TODO: mock database during testing
	// app := fiber.New(fiber.Config{})
	// app.Add(fiber.MethodPost, "/api/chat/:chatID", SendMessage)

	user, err := addRandomUser(DB, false)
	utils.AssertEqual(t, nil, err)

	chat, err := addRandomChatWithNoUsers(DB)
	utils.AssertEqual(t, nil, err)

	err = DB.Model(&chat).Association("Members").Append(&user)
	utils.AssertEqual(t, nil, err, "Chat add Member")
	err = DB.Save(&chat).Error
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
	tx := DB.Find(&message)
	utils.AssertEqual(t, nil, tx.Error)
	utils.AssertEqual(t, message.ID, sendMessageResponse.ID)
	utils.AssertEqual(t, messageContent, message.Content)
	utils.AssertEqual(t, user.ID, message.FromID)
	utils.AssertEqual(t, chat.ID, message.ChatID)
}

func getLoggedInUserSessionCookie(t *testing.T, app *fiber.App, user User) *http.Cookie {
	loginData := LoginRequestSchema{
		Email:    user.Email,
		Password: user.Password,
	}
	b, err := json.Marshal(loginData)
	utils.AssertEqual(t, nil, err)
	buf := bytes.NewReader(b)

	req := httptest.NewRequest(fiber.MethodPost, "/api/login", buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for i := 0; i < len(cookies); i += 1 {
		if cookies[i].Name == SessionIDCookieKey {
			sessionCookie = cookies[i]
		}
	}

	return sessionCookie
}
