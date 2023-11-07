package main

import "github.com/gofiber/fiber/v2"

func setupRoutes(app *fiber.App) {
	app.Get("/", RootHandler)

	api := app.Group("/api")
	ui := app.Group("/ui")

	// TODO: rewrite UI endpoints to use API endpoints internally

	ui.Get("/login", LoginView)
	ui.Post("/login", PostLoginView)
	ui.Get("/chats/:chatID", ChatView)
	ui.Get("/chats", ChatsView)
	ui.Get("/users", UsersView)
	ui.Get("/users/:userID", UserView)
	ui.Get("", HomeView)

	api.Get("/users", GetUsers)
	api.Get("/users/:userID", GetUser)
	api.Post("/users", CreateUser)
	api.Get("/chats", GetChats)
	api.Post("/chats/:chatID", SendMessage)
	api.Post("/users/:userID/avatar", UploadUserAvatar)
	api.Post("/chats/:chatId/users/", JoinChat)
}
