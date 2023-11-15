package main

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func setupRoutes(app *fiber.App) {
	app.Get("/", RootHandler)

	api := app.Group("/api")
	ui := app.Group("/ui")

	// TODO: rewrite UI endpoints to use API endpoints internally

	ui.Get("/login", LoginView)
	ui.Post("/login", PostLoginView)
	ui.Get("/chats", AllChatsView)
	ui.Get("/chats/:chatID", ChatView)
	ui.Get("/users/:userID/chats", UserChatsView)
	ui.Get("/users", UsersView)
	ui.Get("/users/:userID", UserView)
	ui.Get("", HomeView)

	api.Post("/login", Login)
	api.Get("/users", GetUsers)
	api.Get("/users/:userID", GetUser)
	api.Post("/users", CreateUser)
	api.Get("/chats", GetChats)
	api.Get("/chats/:chatID", GetChat)
	api.Post("/chats/:chatID", SendMessage)
	api.Post("/users/:userID/avatar", UploadUserAvatar)
	api.Post("/chats/:chatId/users/", JoinChat)

	app.Get("/ws", websocket.New(WebsocketHandler))
}
