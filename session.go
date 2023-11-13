package main

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type SessionCurrentUser struct {
	ID        uint
	Name      string
	Email     string
	AvatarURL string
}

var SessionCurrentUserKey = "CurrentUser"
var SessionIDCookieKey = "session_id"

func getLoggedInUser(c *fiber.Ctx) (*SessionCurrentUser, error) {
	session, err := store.Get(c)
	if err != nil {
		return nil, err
	}

	val := session.Get(SessionCurrentUserKey)
	if val == nil {
		// TODO: redirect to login page
		return nil, &UnauthorizedUserError{}
	}

	jsonData, ok := val.(string)
	if !ok {
		return nil, errors.New("type casting failed")
	}

	var sessionCurrentUser *SessionCurrentUser
	err = json.Unmarshal([]byte(jsonData), &sessionCurrentUser)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshall")
	}

	// TODO: update fields from db

	return sessionCurrentUser, err
}
