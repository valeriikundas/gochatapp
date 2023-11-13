package main

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name string `validate:"required"`

	Email string `gorm:"uniqueIndex" validate:"required"`
	// TODO: for now without hashing :)
	Password string

	// TODO: add `images` prefix e.g. `images/{filename}.jpg` to this url
	// TODO: use random name for file names
	AvatarURL string

	Chats []Chat `gorm:"many2many:chat_members"`

	Messages []Message `gorm:"foreignKey:FromID"`
}

type Message struct {
	gorm.Model

	Chat   Chat
	ChatID uint `validate:"required"`

	From   User
	FromID uint `validate:"required"`

	Content string `validate:"required"`
}

type Chat struct {
	gorm.Model

	Name    string `gorm:"uniqueIndex" validate:"required"`
	Members []User `gorm:"many2many:chat_members"`

	Messages []Message
}
