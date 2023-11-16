package main

import "gorm.io/gorm"

func getChatUsersExcept(db *gorm.DB, chatID, skipUserID uint) ([]uint, error) {
	var users []User
	tx := db.Joins("JOIN chat_members ON users.id = chat_members.user_id").
		Where("users.id <> ?", skipUserID).
		Where("chat_members.chat_id = ?", chatID).
		Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}

	userIDs := make([]uint, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	return userIDs, nil
}
