package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

// setupTest sets up fiber's App and DB
func setupTest(t *testing.T) (func(), *fiber.App) {
	var clearDB func(*gorm.DB) error
	DB, clearDB = prepareTestDb(t)
	app := createApp(DB)

	return func() {
		err := clearDB(DB)
		if err != nil {
			t.Error(err)
		}
	}, app
}

func clearDB(db *gorm.DB) error {
	tables := []string{"messages", "chat_members", "chats", "users"}
	for _, table := range tables {
		tx := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if tx.Error != nil {
			return tx.Error
		}
	}
	return nil
}

func prepareTestDb(t *testing.T) (*gorm.DB, func(*gorm.DB) error) {
	DB = connectDatabase("chatapp_test")

	err := clearDB(DB)
	if err != nil {
		log.Printf("dropdb failed %v\n", err)
	}

	return DB, clearDB
}

func CreateDBCommand() *exec.Cmd {
	createCommand := fmt.Sprintf("createdb --host %s --port %d --user %s %s", "0.0.0.0", 5432, "valerii", "chatapp_test")
	log.Println(createCommand)
	createCommandSplit := strings.Split(createCommand, " ")
	createCmd := exec.Command(createCommandSplit[0], createCommandSplit[1:]...)
	return createCmd
}

func DropDBCommand() *exec.Cmd {
	command := "dropdb --host 0.0.0.0 --port 5432 --user valerii chatapp_test"
	commandSplit := strings.Split(command, " ")
	cmd := exec.Command(commandSplit[0], commandSplit[1:]...)
	log.Println(cmd)
	return cmd
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
