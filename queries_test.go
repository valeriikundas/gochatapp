package main

import (
	"testing"

	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/utils"
)

func Test_getChatUsersExcept(t *testing.T) {
	t.Skip("currently uses real db")

	config := NewConfig("test_config")

	db := getPostgres(config)

	chat, err := addRandomChatWithUsers(db)
	utils.AssertEqual(t, nil, err)

	chatID := chat.ID

	var chatObj Chat
	tx := db.Model(&Chat{}).Where("id = ?", chatID).Preload("Members").Find(&chatObj)
	utils.AssertEqual(t, nil, tx.Error)

	users := chatObj.Members
	userID := users[0].ID

	userIDs, err := getChatUsersExcept(db, chatID, userID)
	utils.AssertEqual(t, nil, err)

	log.Info("ids=", userIDs)
}
