package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	fiberwebsocket "github.com/gofiber/contrib/websocket"
	"github.com/posener/wstest"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/template/html/v2"
	"github.com/gorilla/websocket"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

func testStatus200(t *testing.T, app *fiber.App, url, method string) {
	t.Helper()

	req := httptest.NewRequest(method, url, nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
}

func testErrorResponse(t *testing.T, err error, resp *http.Response, expectedBodyError string) {
	t.Helper()

	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 500, resp.StatusCode, "Status code")

	body, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, expectedBodyError, string(body), "Response body")
}

func TestGetUsers(t *testing.T) {
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)
	app := createApp(DB)

	users, err := addRandomUsers(DB, 10)
	utils.AssertEqual(t, nil, err)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/users", nil))
	utils.AssertEqual(t, nil, err)

	bytes, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	var data struct {
		Users []User
	}
	json.Unmarshal(bytes, &data)
	utils.AssertEqual(t, len(users), len(data.Users))
	utils.AssertEqual(t, users[0].Name, data.Users[0].Name)
	utils.AssertEqual(t, users[5].Name, data.Users[5].Name)
}

func TestUploadUserAvatar(t *testing.T) {
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)
	app := createApp(DB)

	user, err := addRandomUser(DB, false)
	utils.AssertEqual(t, nil, err)

	fileName := "test.jpeg"
	file, err := os.Open(fileName)
	utils.AssertEqual(t, nil, err)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	formData, err := writer.CreateFormFile("image", fileName)
	utils.AssertEqual(t, nil, err)

	_, err = io.Copy(formData, file)
	utils.AssertEqual(t, nil, err)

	writer.Close()

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/api/users/%d/avatar", user.ID), body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	bytes, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	var data map[string]any
	err = json.Unmarshal(bytes, &data)
	utils.AssertEqual(t, nil, err)

	var resultUser User
	tx := DB.First(&resultUser, user.ID)
	utils.AssertEqual(t, nil, tx.Error)

	// TODO: currently saves into the same repo as prod `uploads`, would be better to make a temporary repo
	utils.AssertEqual(t, fmt.Sprintf("/%s", fileName), resultUser.AvatarURL)
}

func TestChatsView(t *testing.T) {
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)
	app := createApp(DB)

	users, err := addRandomUsers(DB, 10)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 10, len(users))

	_, err = addRandomChatWithUsers(DB)
	utils.AssertEqual(t, nil, err)

	req := httptest.NewRequest(fiber.MethodGet, "/ui/chats", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	// TODO: search chat name and users in html
	// log.Printf("chat=%v\n", chat)
}

func TestGetChatView(t *testing.T) {
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)
	app := createApp(DB)

	users, err := addRandomUsers(DB, 10)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 10, len(users))

	chat, err := addRandomChatWithUsers(DB)
	utils.AssertEqual(t, nil, err)

	user := users[0]
	sessionCookie := getLoggedInUserSessionCookie(t, app, user)

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/ui/chats/%d", chat.ID), nil)
	req.AddCookie(sessionCookie)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	// TODO: test html
}

func TestGetChatViewWithoutWholeApp(t *testing.T) {
	engine := html.New("templates/", ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/base",
	})
	app.Get("/ui/chats/:chatId", ChatView)
	app.Post("/api/login", Login)
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)

	users, err := addRandomUsers(DB, 10)
	utils.AssertEqual(t, nil, err)

	// TODO: test
	// log.Printf("len(users)=%v\n", len(users))

	chat, err := addRandomChatWithUsers(DB)
	utils.AssertEqual(t, nil, err)

	var editedChat Chat
	err = DB.Find(&editedChat, chat.ID).Error
	utils.AssertEqual(t, nil, err)

	user := users[0]
	sessionCookie := getLoggedInUserSessionCookie(t, app, user)

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/ui/chats/%d", chat.ID), nil)
	req.AddCookie(sessionCookie)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test()")

	_, err = io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	_, err = io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	// TODO: test
	// log.Printf("%v\n", string(bytes))
}

func TestJoinChat(t *testing.T) {
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)

	app := createApp(DB)

	user, err := addRandomUser(DB, false)
	utils.AssertEqual(t, nil, err)

	chat, err := addRandomChatWithNoUsers(DB)
	utils.AssertEqual(t, nil, err)

	chatsLenInitial := len(user.Chats)
	data := struct {
		Email string
	}{
		Email: user.Email,
	}

	jsonData, err := json.Marshal(data)
	utils.AssertEqual(t, nil, err)

	body := bytes.NewReader(jsonData)
	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/api/chats/%d/users", chat.ID), body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	var editedUser User
	err = DB.Preload("Chats").Find(&editedUser, user.ID).Error
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, chatsLenInitial+1, len(editedUser.Chats), "len Chats did not change")

	isChatFound := false
	for _, chatIter := range editedUser.Chats {
		if chatIter.ID == chat.ID {
			isChatFound = true
			break
		}
	}
	utils.AssertEqual(t, true, isChatFound)
}

func TestSendMessageToWebsocket(t *testing.T) {
	t.Skip("websockets are not implemented")

	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)
	app := createApp(DB)

	users, err := addRandomUsers(DB, 10)
	utils.AssertEqual(t, nil, err)

	chat, err := addRandomChatWithNoUsers(DB)
	utils.AssertEqual(t, nil, err)

	chatId := chat.ID

	// join chat
	user := users[0].Email
	data := struct {
		Email string
	}{
		Email: user,
	}
	jsonData, err := json.Marshal(data)
	utils.AssertEqual(t, nil, err)
	body := bytes.NewReader(jsonData)
	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/api/chats/%d/users", chatId), body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	// send message to chat
	testServer := httptest.NewServer(adaptor.FiberHandlerFunc(fiberwebsocket.New(WebsocketHandler)))
	defer testServer.Close()

	// url := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws"
	// conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	// utils.AssertEqual(t, nil, err)
	// defer conn.Close()
	// utils.AssertEqual(t, fiber.StatusSwitchingProtocols, resp.StatusCode)
	// b, err := io.ReadAll(resp.Body)
	// utils.AssertEqual(t, nil, err)
	// log.Printf("log response body1=%s\n", string(b))

	// log.Printf("conn.LocalAddr()=%s", conn.LocalAddr())

	url := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws/5"
	log.Printf("log url=%s\n", url)
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	utils.AssertEqual(t, nil, err)
	defer conn.Close()
	utils.AssertEqual(t, fiber.StatusSwitchingProtocols, resp.StatusCode, "Status code")

	// b, err := io.ReadAll(resp.Body)
	// utils.AssertEqual(t, nil, err)
	// log.Printf("log response body2=%s\n", string(b))

	// v := map[string]any{
	// 	"a": "b",
	// 	"x": 123,
	// }
	// jsonData, err = json.Marshal(v)
	// utils.AssertEqual(t, nil, err)
	// log.Printf(" conn.WriteMessage\n")
	// err = conn.WriteMessage(websocket.TextMessage, jsonData)
	// utils.AssertEqual(t, nil, err, "websocket WriteMessage failed")

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("%v", err)
		}
		_, p, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}

	// test message was created in database
}

var upgrader = websocket.Upgrader{}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("websocket read message mt=%d msg=%s\n", mt, message)

		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func TestExample(t *testing.T) {
	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("%v", err)
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}
}

// https://stackoverflow.com/questions/47637308/create-unit-test-for-ws-in-golang

func TestHandler(t *testing.T) {
	t.Skip()

	dialer := wstest.NewDialer(adaptor.FiberHandler(fiberwebsocket.New(WebsocketHandler)))
	url := "ws://whatever/ws"
	conn, resp, err := dialer.Dial(url, nil)
	utils.AssertEqual(t, nil, err)
	defer conn.Close()
	utils.AssertEqual(t, fiber.StatusSwitchingProtocols, resp.StatusCode)

	data := map[string]any{
		"hello": "world",
	}
	v, err := json.Marshal(data)
	utils.AssertEqual(t, nil, err)
	log.Printf("will write json\n")
	err = conn.WriteJSON(v)
	log.Printf("finished write json\n")
	utils.AssertEqual(t, nil, err)
}
