package main

import (
	"log"
	"testing"

	"github.com/gofiber/fiber/v2/utils"
)

func Test_getChatUsersExcept(t *testing.T) {
	t.Skip("currently uses real db")

	DB = connectDatabase("chatapp")

	chat, err := addRandomChatWithUsers(DB)
	utils.AssertEqual(t, nil, err)

	chatID := chat.ID

	var chatObj Chat
	tx := DB.Model(&Chat{}).Where("id = ?", chatID).Preload("Members").Find(&chatObj)
	utils.AssertEqual(t, nil, tx.Error)

	users := chatObj.Members
	userID := users[0].ID

	userIDs, err := getChatUsersExcept(chatID, userID)
	utils.AssertEqual(t, nil, err)

	log.Println("ids=", userIDs)
}
