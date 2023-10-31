package main

import (
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

func addRandomUser(db *gorm.DB) (*User, error) {
	user := User{
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}
	tx := db.Create(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func addRandomChat(db *gorm.DB) (*Chat, error) {
	var users []User
	tx := db.Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}

	mn := int(math.Min(float64(len(users)), float64(10)))
	cnt := rand.Intn(mn)
	chat := Chat{
		Name:    strings.Join([]string{gofakeit.MovieName(), gofakeit.BookTitle(), gofakeit.Country()}, " - "),
		Members: selectRandomUsers(users, cnt),
	}
	tx = db.Create(&chat)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &chat, nil
}

func generateRandomChats(t *testing.T, db *gorm.DB) error {
	userData := make([]User, 100)
	for i := 0; i < 100; i += 1 {
		user, err := addRandomUser(db)
		if err != nil {
			return err
		}
		userData[i] = *user
	}
	tx := db.CreateInBatches(userData, 10)
	if t == nil {
		if tx.Error != nil {
			return tx.Error
		}
	} else {
		utils.AssertEqual(t, nil, tx.Error)
	}

	var users []User
	tx = db.Find(&users)
	if t == nil {
		if tx.Error != nil {
			return tx.Error
		}
	} else {
		utils.AssertEqual(t, int64(100), tx.RowsAffected)
		utils.AssertEqual(t, nil, tx.Error)
	}

	k := 100
	chatData := make([]Chat, k)
	for i := 0; i < k; i += 1 {
		chat, err := addRandomChat(db)
		if err != nil {
			return err
		}
		chatData[i] = *chat
	}
	tx = db.CreateInBatches(chatData, 10)
	if t == nil {
		if tx.Error != nil {
			return tx.Error
		}
	} else {
		utils.AssertEqual(t, nil, tx.Error)
	}

	for _, chat := range chatData {
		messagesCount := rand.Intn(100)
		messages := make([]Message, messagesCount)
		for i := 0; i < messagesCount; i += 1 {
			userIndex := rand.Intn(len(userData))
			fromID := userData[userIndex].ID
			messageLength := rand.Intn(20)
			messages[i] = Message{
				ChatID:  chat.ID,
				FromID:  fromID,
				Content: gofakeit.LoremIpsumSentence(messageLength),
			}
		}
		tx := db.Create(&messages)
		if t == nil {
			if tx.Error != nil {
				return tx.Error
			}
		} else {
			utils.AssertEqual(t, nil, tx.Error)
		}
	}

	return nil
}

func selectRandomUsers(users []User, cnt int) []User {
	n := len(users)
	randomIndices := make([]int, n)
	for i := range randomIndices {
		randomIndices[i] = i
	}
	rand.Shuffle(n, func(i, j int) {
		randomIndices[j], randomIndices[i] = randomIndices[i], randomIndices[j]
	})
	selectedUsers := make([]User, cnt)
	for i := 0; i < cnt; i += 1 {
		selectedUsers[i] = users[randomIndices[i]]
	}
	return selectedUsers
}
