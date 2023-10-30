package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/driver/postgres"
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
	db, dropCmd := prepareTestDb(t)
	defer dropCmd.Run()
	app := createApp(db)

	generateRandomChats(t, db)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/chats", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	var data ChatsResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 100, len(data.Chats))
}

func prepareTestDb(t *testing.T) (db *gorm.DB, dropCmd *exec.Cmd) {
	dropCmd = createDropDbCommand()
	err := dropCmd.Run()
	if err != nil {
		log.Printf("warning dropdb failed %v\n", err)
	}

	createCommand := "createdb --port 5432 --user valerii chatapp_test"
	createCommandSplit := strings.Split(createCommand, " ")
	createCmd := exec.Command(createCommandSplit[0], createCommandSplit[1:]...)
	bytes, err := createCmd.Output()
	utils.AssertEqual(t, nil, err, fmt.Sprintf("bytes %v", bytes))

	dsn := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable TimeZone=Europe/Kiev", "0.0.0.0", 5432, "chatapp_test")
	db, err = gorm.Open(postgres.Open(dsn))
	utils.AssertEqual(t, nil, err)

	err = db.AutoMigrate(&User{})
	utils.AssertEqual(t, nil, err)

	dropCmd = createDropDbCommand()
	return
}

func createDropDbCommand() *exec.Cmd {
	command := "dropdb --port 5432 --user valerii chatapp_test"
	commandSplit := strings.Split(command, " ")
	cmd := exec.Command(commandSplit[0], commandSplit[1:]...)
	return cmd
}
