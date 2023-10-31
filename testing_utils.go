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
	dropCmd = dropDBCommand()
	err := dropCmd.Run()
	if err != nil {
		log.Printf("dropdb failed %v\n", err)
	}

	createCmd := createDBCommand()
	bytes, err := createCmd.Output()
	utils.AssertEqual(t, nil, err, fmt.Sprintf("bytes %v", bytes))

	db = connectDatabase("chatapp_test")
	dropCmd = dropDBCommand()
	return
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
