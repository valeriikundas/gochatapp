package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

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

	db = connectDatabase("chatapp_test")
	dropCmd = createDropDbCommand()
	return
}

func createDropDbCommand() *exec.Cmd {
	command := "dropdb --port 5432 --user valerii chatapp_test"
	commandSplit := strings.Split(command, " ")
	cmd := exec.Command(commandSplit[0], commandSplit[1:]...)
	return cmd
}
