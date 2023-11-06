package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

func ChatsView(c *fiber.Ctx) error {
	var chats []Chat
	tx := DB.Model(&Chat{}).Preload("Members").Find(&chats)
	if tx.Error != nil {
		return tx.Error
	}
	return c.Render("chats", fiber.Map{
		"Chats": chats,
	})
}

func UsersView(c *fiber.Ctx) error {
	users, err := getUsers()
	if err != nil {
		return err
	}

	usersResponse := make([]UserResponse, len(users))
	for i := 0; i < len(users); i += 1 {
		usersResponse[i] = getUserResponse(users[i])
	}

	return c.Render("users", fiber.Map{
		"Users": usersResponse,
	})
}

func UserView(c *fiber.Ctx) error {
	var user User
	userID, err := c.ParamsInt("userID")
	if err != nil {
		return err
	}
	// TODO: load members only for one queried record
	tx := DB.Preload("Chats").Preload("Chats.Members").First(&user, userID)
	if tx.Error != nil {
		return tx.Error
	}

	return c.Render("user", fiber.Map{
		"User": getUserResponse(user),
	})
}

// TODO: don't like this name
func getUserResponse(user User) UserResponse {
	var avatarURL string
	if user.AvatarFileName != "" {
		log.Printf("avatarFilePath=%v\n", user.AvatarFileName)
		avatarURL = fmt.Sprintf("/%s", user.AvatarFileName)
		log.Printf("avatarURL=%v\n", avatarURL)
	} else {
		avatarURL = ""
	}

	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: avatarURL,
	}
}

func getUsersResponse(users []User) []UserResponse {
	usersResponse := make([]UserResponse, len(users))
	for i := 0; i < len(users); i += 1 {
		usersResponse[i] = getUserResponse(users[i])
	}
	return usersResponse
}

func ChatView(c *fiber.Ctx) error {
	chatID, err := c.ParamsInt("chatId", -1)
	if err != nil {
		return err
	}
	if chatID == -1 {
		return errors.New("chatId param missing in URL")
	}
	var chat Chat
	tx := DB.Preload("Members").Where("id = ?", chatID).Preload("Messages.From").First(&chat)
	if tx.Error != nil {
		return tx.Error
	}

	chatResponse := ChatResponse{
		Name:    chat.Name,
		Members: getUsersResponse(chat.Members),
	}

	var user User
	// todo: implement current user functionality
	tx = DB.Take(&user)
	if tx.Error != nil {
		return tx.Error
	}

	// FIXME: if I pass `User` but with other fields and `layout` present, it
	// does not throw an error, but it should. needs deeper look into fiber
	// source code
	return c.Render("chat", fiber.Map{
		"Chat": chatResponse,
		"User": user,
	})

	// var buf bytes.Buffer
	// tmpl := template.Must(template.ParseFiles("templates/chat.html"))
	// data := fiber.Map{
	// 	"Chat": chat,
	// 	"User": user,
	// }
	// err = tmpl.Execute(&buf, data)
	// if err != nil {
	// 	return err
	// }

	// bytes, err := io.ReadAll(&buf)
	// if err != nil {
	// 	return err
	// }

	// body := string(bytes)

	// return c.SendString(body)
}

func HomeView(c *fiber.Ctx) error {
	return c.Render("home", fiber.Map{
		"a": "b",
	})
}

type UsersResponse struct {
	Users []UserResponse
}

func getUsers() ([]User, error) {
	var users []User
	tx := DB.Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return users, nil
}

func GetUsers(c *fiber.Ctx) error {
	users, err := getUsers()
	if err != nil {
		return err
	}

	usersResponse := make([]UserResponse, len(users))
	for i := 0; i < len(users); i += 1 {
		usersResponse[i] = getUserResponse(users[i])
	}

	data := UsersResponse{
		Users: usersResponse,
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	// TODO: return only requested fields, no created_at,deleted_at,messages etc for all route handlers
	return c.SendString(string(bytes))
}

type UserResponse struct {
	ID        uint
	Name      string
	Email     string
	AvatarURL string
}

type ChatResponse struct {
	Name     string
	Members  []UserResponse
	Messages []Message
}

func GetUser(c *fiber.Ctx) error {
	var user User
	userID, err := c.ParamsInt("userID")
	if err != nil {
		return err
	}
	tx := DB.Preload("Chats").First(&user, userID)
	if tx.Error != nil {
		return tx.Error
	}

	log.Printf("avatarFilePath=%v\n", user.AvatarFileName)
	avatarURL := fmt.Sprintf("/images/%s", user.AvatarFileName)
	log.Printf("avatarURL=%v\n", avatarURL)

	userResponse := UserResponse{
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: avatarURL,
	}

	response := map[string]any{
		"User": userResponse,
	}
	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	return c.SendString(string(bytes))
}

type FieldError struct {
	Field, Tag, Param string
}

func CreateUser(c *fiber.Ctx) error {
	var user User
	err := c.BodyParser(&user)
	if err != nil {
		return err
	}

	err = validate.Struct(user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	// if err != nil {
	// 	var errors []FieldError
	// 	for _, err := range err.(validator.ValidationErrors) {
	// 		el := FieldError{
	// 			Field: err.Field(),
	// 			Tag:   err.Tag(),
	// 			Param: err.Param(),
	// 		}
	// 		errors = append(errors, el)
	// 	}
	// 	return c.Status(fiber.StatusBadRequest).JSON(errors)
	// }
	tx := DB.Create(&user)
	if tx.Error != nil {
		return tx.Error
	}
	return c.JSON(user)
}

type GetChatsResponse struct {
	Chats []Chat
}

func GetChats(c *fiber.Ctx) error {
	var chats []Chat
	ch := &Chat{}
	model := DB.Model(ch)
	query := model.Preload("Members")
	tx := query.Find(&chats)
	if tx.Error != nil {
		return tx.Error
	}

	data := GetChatsResponse{
		Chats: chats,
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return c.SendString(string(bytes))
}

type SendMessageRequest struct {
	FromID  uint
	ChatID  uint
	Content string
}

func SendMessage(c *fiber.Ctx) error {
	var data SendMessageRequest
	err := c.BodyParser(&data)
	if err != nil {
		return errors.Wrap(err, "BodyParser failed")
	}

	err = validate.Struct(data)
	if err != nil {
		var errors []FieldError
		for _, err := range err.(validator.ValidationErrors) {
			el := FieldError{
				Field: err.Field(),
				Tag:   err.Tag(),
				Param: err.Param(),
			}
			errors = append(errors, el)
		}
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	message := Message{
		ChatID:  data.ChatID,
		FromID:  data.FromID,
		Content: data.Content,
	}
	tx := DB.Create(&message)

	if tx.Error != nil {
		logger.Errorf("er=%v\n", tx.Statement.Error)
		return errors.Wrap(tx.Error, "db create message failed")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ID": message.ID,
	})
}

func RootHandler(c *fiber.Ctx) error {
	return c.Redirect("/ui", fiber.StatusPermanentRedirect)
}

func UploadUserAvatar(c *fiber.Ctx) error {
	log.Println("UploadUserAvatar")

	file, err := c.FormFile("image")
	if err != nil {
		return err
	}

	_, err = os.Stat("uploads/")
	if os.IsNotExist(err) {
		os.MkdirAll("./uploads", 0744)
	}

	fileName := file.Filename
	// filePath, err := url.JoinPath("uploads", fileName)
	// if err != nil {
	// 	return err
	// }
	filePath := fmt.Sprintf("uploads/%s", fileName)
	err = c.SaveFile(file, filePath)
	if err != nil {
		log.Printf("err=%v\n", err)
		return err
	}

	userID := c.Params("userID")
	tx := DB.Model(&User{}).Where("id = ?", userID).Update("AvatarFileName", fileName)
	if tx.Error != nil {
		return tx.Error
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
