package main

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

func generateRandomChats(t *testing.T, db *gorm.DB) error {
	userData := make([]User, 100)
	for i := 0; i < 100; i += 1 {
		userData[i] = User{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
		}
	}
	tx := db.CreateInBatches(userData, 10)
	if t == nil {
		return tx.Error
	} else {
		utils.AssertEqual(t, nil, tx.Error)
	}

	var users []User
	tx = db.Find(&users)
	if t == nil {
		return tx.Error
	} else {
		utils.AssertEqual(t, int64(100), tx.RowsAffected)
		utils.AssertEqual(t, nil, tx.Error)
	}

	k := 100
	chatData := make([]Chat, k)
	for i := 0; i < k; i += 1 {
		cnt := rand.Intn(10)
		chatData[i] = Chat{
			Name:    strings.Join([]string{gofakeit.MovieName(), gofakeit.BookTitle(), gofakeit.Country()}, " - "),
			Members: selectRandomUsers(users, cnt),
		}
	}
	chatNames := make([]string, len(chatData))
	for i := range chatNames {
		chatNames[i] = chatData[i].Name
	}
	tx = db.CreateInBatches(chatData, 10)
	if t == nil {
		return tx.Error
	} else {
		utils.AssertEqual(t, nil, tx.Error)
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
