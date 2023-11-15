package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
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

func createDBCommand() *exec.Cmd {
	createCommand := fmt.Sprintf("createdb --host %s --port %d --user %s %s", "0.0.0.0", 5432, "valerii", "chatapp_test")
	log.Println(createCommand)
	createCommandSplit := strings.Split(createCommand, " ")
	createCmd := exec.Command(createCommandSplit[0], createCommandSplit[1:]...)
	return createCmd
}

func dropDBCommand() *exec.Cmd {
	command := "dropdb --host 0.0.0.0 --port 5432 --user valerii chatapp_test"
	commandSplit := strings.Split(command, " ")
	cmd := exec.Command(commandSplit[0], commandSplit[1:]...)
	log.Println(cmd)
	return cmd
}
