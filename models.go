package main

import "gorm.io/gorm"

type User struct {
	*gorm.Model

	Name  string `gorm:"uniqueIndex" validate:"required"`
	Email string `gorm:"uniqueIndex" validate:"required"`

	Chats []Chat `gorm:"many2many:chat_members"`
}

type Chat struct {
	*gorm.Model

	Name    string `gorm:"uniqueIndex" validate:"required"`
	Members []User `gorm:"many2many:chat_members"`
}
