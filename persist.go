package main

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func saveMessage(db *gorm.DB, userEmail string, chatID uint, messageContent string) (uint, error) {
	var user User
	tx := db.Where("Email = ?", userEmail).First(&user)
	if tx.Error != nil {
		return 0, tx.Error
	}

	message := Message{
		ChatID:  chatID,
		FromID:  user.ID,
		Content: messageContent,
	}
	tx = db.Create(&message)
	if tx.Error != nil {
		return 0, errors.Wrap(tx.Error, "db create message failed")
	}

	return message.ID, nil
}
