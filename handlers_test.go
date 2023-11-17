package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	fiberwebsocket "github.com/gofiber/contrib/websocket"
	"github.com/posener/wstest"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gorilla/websocket"
)

func testStatus200(t *testing.T, app *fiber.App, url, method string) []byte {
	t.Helper()

	req := httptest.NewRequest(method, url, nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")

	b, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	return b
}

func TestGetUsers(t *testing.T) {
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

	users, err := addRandomUsers(DB, 10)
	utils.AssertEqual(t, nil, err)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/users", nil))
	utils.AssertEqual(t, nil, err)

	bytes, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	var data struct {
		Users []User
	}
	err = json.Unmarshal(bytes, &data)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, len(users), len(data.Users))
	utils.AssertEqual(t, users[0].Name, data.Users[0].Name)
	utils.AssertEqual(t, users[5].Name, data.Users[5].Name)
}

func TestUploadUserAvatar(t *testing.T) {
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

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
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

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
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

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
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

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
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

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

	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

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
	log.Debugf("log url=%s\n", url)
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

// func testWebsocketHandler(w http.ResponseWriter, r *http.Request) {
// 	c, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		return
// 	}
// 	defer c.Close()

// 	for {
// 		mt, message, err := c.ReadMessage()
// 		if err != nil {
// 			break
// 		}
// 		log.Printf("websocket read message mt=%d msg=%s\n", mt, message)

// 		err = c.WriteMessage(mt, message)
// 		if err != nil {
// 			break
// 		}
// 	}
// }

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
	log.Debug("will write json\n")
	err = conn.WriteJSON(v)
	log.Debug("finished write json\n")
	utils.AssertEqual(t, nil, err)
}

func TestPostLoginViewWithExistingUser(t *testing.T) {
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

	user, err := addRandomUser(DB, false)
	utils.AssertEqual(t, nil, err)

	v := LoginRequestSchema{
		Email:    user.Email,
		Password: user.Password,
	}
	b, err := json.Marshal(v)
	utils.AssertEqual(t, nil, err)
	body := bytes.NewReader(b)
	loginReq := httptest.NewRequest(fiber.MethodPost, "/ui/login", body)
	loginReq.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(loginReq)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionIDCookieKey {
			sessionCookie = c
		}
	}

	if sessionCookie == nil {
		t.Error("empty session cookie")
	}
}

func TestPostLoginViewWithNonExistingUser(t *testing.T) {
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

	email := "test@test.com"
	password := "pass"
	v := LoginRequestSchema{
		Email:    email,
		Password: password,
	}
	b, err := json.Marshal(v)
	utils.AssertEqual(t, nil, err)
	body := bytes.NewReader(b)
	loginReq := httptest.NewRequest(fiber.MethodPost, "/ui/login", body)
	loginReq.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(loginReq)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionIDCookieKey {
			sessionCookie = c
		}
	}

	if sessionCookie == nil {
		t.Error("empty session cookie")
	}

	var createdUser User
	err = DB.Where("Email = ?", email).Find(&createdUser).Error
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, email, createdUser.Email)
	utils.AssertEqual(t, password, createdUser.Password)
}

func TestRenderUsers(t *testing.T) {
	app, db, teardownTest := setupTest(t)
	defer teardownTest()

	users, err := addRandomUsers(db, 10)
	utils.AssertEqual(t, nil, err)

	req := httptest.NewRequest(fiber.MethodGet, "/ui/users", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	if resp.StatusCode != fiber.StatusOK {
		b, err := io.ReadAll(resp.Body)
		utils.AssertEqual(t, nil, err)
		t.Errorf("Status code=%d, body=%v\n", resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	html := string(b)
	utils.AssertEqual(t, len(users), strings.Count(html, "user-row"))
}

func TestUserChatsView(t *testing.T) {
	app, db, teardownTest := setupTest(t)
	defer teardownTest()

	user, err := addRandomUser(db, false)
	utils.AssertEqual(t, nil, err)

	chatsCount := 9
	_, err = addRandomChatsForUser(db, *user, chatsCount)
	utils.AssertEqual(t, nil, err)

	userID := user.ID

	cookie := getLoggedInUserSessionCookie(t, app, *user)

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/ui/users/%d/chats", userID), nil)
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	cnt := strings.Count(string(b), "chat-row")
	utils.AssertEqual(t, chatsCount, cnt)
}

func TestUserView(t *testing.T) {
	app, db, teardownTest := setupTest(t)
	defer teardownTest()

	user, err := addRandomUser(db, false)
	utils.AssertEqual(t, nil, err)

	chats, err := addRandomChatsForUser(db, *user, 5)
	utils.AssertEqual(t, nil, err)

	userID := user.ID
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/ui/users/%d", userID), nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	html := string(b)

	utils.AssertEqual(t, true, strings.Contains(html, user.Name))
	utils.AssertEqual(t, true, strings.Contains(html, user.Email))
	utils.AssertEqual(t, len(chats), strings.Count(html, "chat-row"))
}

func TestHomeView(t *testing.T) {
	app, _, teardownTest := setupTest(t)
	defer teardownTest()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ui", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}

func TestRootHandler(t *testing.T) {
	app, _, teardownTest := setupTest(t)
	defer teardownTest()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusPermanentRedirect, resp.StatusCode)
}

func TestGetUser(t *testing.T) {
	app, db, teardownTest := setupTest(t)
	defer teardownTest()

	user, err := addRandomUser(db, false)
	utils.AssertEqual(t, nil, err)

	chats, err := addRandomChatsForUser(db, *user, 5)
	utils.AssertEqual(t, nil, err)

	userID := user.ID

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/api/users/%d", userID), nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "application/json", resp.Header.Get("Content-Type"))

	b, err := io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	var v struct {
		User User
	}
	err = json.Unmarshal(b, &v)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, user.Name, v.User.Name)
	utils.AssertEqual(t, user.Email, v.User.Email)
	utils.AssertEqual(t, len(chats), len(v.User.Chats))
}

func TestCreateUser(t *testing.T) {
	app, _, teardownTest := setupTest(t)
	defer teardownTest()

	userToCreate := User{
		Name:  "test",
		Email: "test@gmail.com",
	}

	b, err := json.Marshal(userToCreate)
	utils.AssertEqual(t, nil, err)

	body := bytes.NewReader(b)

	req := httptest.NewRequest(fiber.MethodPost, "/api/users", body)
	req.Header.Add("Content-Type", "application/json")
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	b, err = io.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	var v User
	err = json.Unmarshal(b, &v)
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, userToCreate.Name, v.Name)
	utils.AssertEqual(t, userToCreate.Email, v.Email)
}

func TestGetChat(t *testing.T) {
	app, DB, teardownTest := setupTest(t)
	defer teardownTest()

	_, err := addRandomUsers(DB, 20)
	utils.AssertEqual(t, nil, err)

	chat, err := addRandomChatWithUsers(DB)
	utils.AssertEqual(t, nil, err)

	chatID := chat.ID

	b := testStatus200(t, app, fmt.Sprintf("/api/chats/%d", chatID), fiber.MethodGet)

	var v struct {
		Chat Chat
	}
	err = json.Unmarshal(b, &v)
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, chat.ID, v.Chat.ID)
	utils.AssertEqual(t, chat.Name, v.Chat.Name)
	utils.AssertEqual(t, len(chat.Members), len(v.Chat.Members))
}
