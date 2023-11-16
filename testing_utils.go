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
func setupTest(t *testing.T) (*fiber.App, *gorm.DB, func()) {
	config := NewConfig("test_config")
	db, clearDB := prepareTestDb(t, config)

	// TODO: write test for session store

	redisDB := getRedis(config)

	app := createApp(db, redisDB)

	return app, db, func() {
		// TODO: errors in this func should not affect next functions. how to do that?

		err := clearDB(db)
		if err != nil {
			t.Error(err)
		}

		err = redisDB.Reset()
		if err != nil {
			t.Error(err)
		}
	}
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

func prepareTestDb(t *testing.T, config *Config) (*gorm.DB, func(*gorm.DB) error) {
	db := getPostgres(config)

	err := clearDB(db)
	if err != nil {
		log.Printf("dropdb failed %v\n", err)
	}

	return db, clearDB
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
