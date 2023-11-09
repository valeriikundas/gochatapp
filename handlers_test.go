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
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/template/html/v2"
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

	user := users[0]
	req := httptest.NewRequest(fiber.MethodGet, "/ui/chats", nil)
	req.AddCookie(&http.Cookie{
		Name:  "Authorization",
		Value: user.Email,
		Path:  "/",
	})
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

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

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/ui/chats/%d", chat.ID), nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	// TODO: test html
}

func TestGetChatViewWithoutWholeApp(t *testing.T) {
	engine := html.New("templates/", ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/base",
	})
	app.Get("/:chatId", ChatView)
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	defer clearDB(DB)

	users, err := addRandomUsers(DB, 10)
	utils.AssertEqual(t, nil, err)

	// TODO: test
	log.Printf("%v\n", len(users))

	chat, err := addRandomChatWithUsers(DB)
	utils.AssertEqual(t, nil, err)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/%d", chat.ID), nil))
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
