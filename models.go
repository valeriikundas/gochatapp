package main

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name      string `gorm:"uniqueIndex" validate:"required"`
	Email     string `gorm:"uniqueIndex" validate:"required"`
	AvatarURL string

	Chats []Chat `gorm:"many2many:chat_members"`
}

type Message struct {
	gorm.Model

	Chat   Chat
	ChatID uint

	From   User `validate:"required"`
	FromID uint

	Content string `validate:"required"`
}

type Chat struct {
	gorm.Model

	Name    string `gorm:"uniqueIndex" validate:"required"`
	Members []User `gorm:"many2many:chat_members"`

	Messages []Message
}
