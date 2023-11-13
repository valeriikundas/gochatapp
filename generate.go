package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

func addRandomUser(db *gorm.DB, addAvatar bool) (*User, error) {
	username := gofakeit.Name()

	var avatarURL string
	if addAvatar {
		cleanedUsername := strings.ReplaceAll(username, " ", "")
		fileName := fmt.Sprintf("%s.jpg", cleanedUsername)
		URL := "https://random.imagecdn.app/300/200"
		err := loadImageFromURL(URL, fileName)
		if err != nil {
			return nil, err
		}

		avatarURL = fmt.Sprintf("/%s", fileName)
	} else {
		avatarURL = ""
	}

	password := gofakeit.Password(true, true, true, true, true, 20)

	user := User{
		Name:      username,
		Password:  password,
		Email:     gofakeit.Email(),
		AvatarURL: avatarURL,
	}
	tx := db.Create(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func loadImageFromURL(URL, fileName string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filePath := fmt.Sprintf("uploads/%s", fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func addRandomUsers(db *gorm.DB, n int) ([]User, error) {
	users := make([]User, n)
	for i := 0; i < n; i += 1 {
		users[i] = User{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
		}
	}
	tx := db.Create(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return users, nil
}

func addRandomChatWithNoUsers(db *gorm.DB) (*Chat, error) {
	chat := Chat{
		Name: strings.Join([]string{gofakeit.MovieName(), gofakeit.BookTitle(), gofakeit.Country()}, " - "),
	}
	tx := db.Create(&chat)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &chat, nil
}

func addRandomChatWithUsers(db *gorm.DB) (*Chat, error) {
	var users []User
	tx := db.Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if len(users) == 0 {
		return nil, errors.New("cannot add random chat as there are no users in the database")
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
	// TODO: refactor to focused functions

	users, err := addRandomUsers(DB, 100)
	utils.AssertEqual(t, nil, err)

	k := 100
	chatData := make([]Chat, k)
	for i := 0; i < k; i += 1 {
		chat, err := addRandomChatWithUsers(db)
		if t == nil {
			if err != nil {
				return err
			}
		} else {
			utils.AssertEqual(t, nil, err)
		}
		chatData[i] = *chat
	}

	for _, chat := range chatData {
		messagesCount := rand.Intn(50) + 1
		messages := make([]Message, messagesCount)
		for i := 0; i < messagesCount; i += 1 {
			userIndex := rand.Intn(len(users))
			fromID := users[userIndex].ID
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
