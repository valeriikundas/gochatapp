package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// TODO: split ui and api handlers
// TODO: replace `log` with `github.com/gofiber/fiber/v2/log`

func AllChatsView(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	sessionCurrentUser, err := getLoggedInUser(c)
	_, isUnauthorizedUserError := err.(*UnauthorizedUserError)
	if err != nil {
		if !isUnauthorizedUserError {
			return errors.Wrap(err, "getLoggedInUser")
		}
	}

	var chats []Chat
	tx := db.Model(&Chat{}).Preload("Members").Find(&chats)
	if tx.Error != nil {
		return tx.Error
	}

	var user *User
	if sessionCurrentUser != nil {
		userEmail := sessionCurrentUser.Email

		tx = db.Where("Email = ?", userEmail).First(&user)
		if tx.Error != nil {
			return errors.Wrap(tx.Error, "User not found")
		}
	}

	return c.Render("chats", fiber.Map{
		"Chats":       chats,
		"Mode":        "all",
		"CurrentUser": user,
	})
}

func UserChatsView(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	sessionCurrentUser, err := getLoggedInUser(c)
	if err != nil {
		return err
	}

	userEmail := sessionCurrentUser.Email

	var user User
	err = db.Preload("Chats.Members").Where("Email = ?", userEmail).First(&user).Error
	if err != nil {
		return errors.Wrap(err, "Get user by email")
	}

	userChats := user.Chats

	return c.Render("chats", fiber.Map{
		"Chats":       userChats,
		"Mode":        "joined",
		"CurrentUser": *sessionCurrentUser,
	})
}

func UsersView(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	users, err := getUsers(db)
	if err != nil {
		return err
	}

	sessionCurrentUser, err := getLoggedInUser(c)
	if err != nil {
		_, isUnauthorizedUserError := err.(*UnauthorizedUserError)
		if !isUnauthorizedUserError {
			return err
		}
	}

	return c.Render("users", fiber.Map{
		"Users":       users,
		"CurrentUser": sessionCurrentUser,
	})
}

func UserView(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	var user User
	userID, err := c.ParamsInt("userID")
	if err != nil {
		return err
	}
	// TODO: load members only for one queried record
	tx := db.Preload("Chats").Preload("Chats.Members").First(&user, userID)
	if tx.Error != nil {
		return tx.Error
	}

	return c.Render("user", fiber.Map{
		"User": user,
	})
}

func ChatView(c *fiber.Ctx) error {
	// TODONEXT:
	// TODO: select current user feature
	// TODO: disallow sending message if user has not joined the chat?
	// TODO: join chat feature
	// TODO: send message feature
	// TODO: leave chat feature
	// TODO: bots that talk live

	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	sessionCurrentUser, err := getLoggedInUser(c)
	if err != nil {
		return errors.Wrap(err, "getLoggedInUser")
	}

	chatID, err := c.ParamsInt("chatID", -1)
	if err != nil {
		return errors.Wrap(err, "ParamsInt")
	}
	if chatID == -1 {
		return errors.New("chatID param missing in URL")
	}
	var chat Chat
	tx := db.Preload("Members").Where("id = ?", chatID).Preload("Messages.From").First(&chat)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "get chat by id")
	}

	var user *User
	if sessionCurrentUser != nil {
		// todo: implement current user functionality
		tx = db.Where("Email = ?", sessionCurrentUser.Email).First(&user)
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return errors.Wrap(tx.Error, "User not found")
		} else if tx.Error != nil {
			return tx.Error
		}
	}

	// FIXME: if I pass `User` but with other fields and `layout` present, it
	// does not throw an error, but it should. needs deeper look into fiber
	// source code
	return c.Render("chat", fiber.Map{
		"Chat":        chat,
		"CurrentUser": user,
	})

	// NOTE: below is a code that makes failing template realy fail

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
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	sessionCurrentUser, err := getLoggedInUser(c)
	if err != nil {
		_, isUnauthorizedUserError := err.(*UnauthorizedUserError)
		if isUnauthorizedUserError {
			//
		} else {
			return errors.Wrap(err, "getLoggedInUser()")
		}
	}

	var currentUser *User
	if sessionCurrentUser != nil {
		userEmail := sessionCurrentUser.Email
		err = db.Where("Email = ?", userEmail).First(&currentUser).Error
		if err != nil {
			return errors.Wrap(err, "Get user by email")
		}
	}

	return c.Render("home", fiber.Map{
		"CurrentUser": currentUser,
	})
}

func getUsers(db *gorm.DB) ([]User, error) {
	var users []User
	tx := db.Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return users, nil
}

type LoginRequestSchema struct {
	Email    string
	Password string
}

func Login(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	store, ok := c.Locals("store").(*session.Store)
	if !ok {
		log.Fatalf("error getting `store` from c.Locals()")
	}

	session, err := store.Get(c)
	if err != nil {
		return err
	}

	// TODO: implement passwords
	var loginData LoginRequestSchema
	err = c.BodyParser(&loginData)
	if err != nil {
		return errors.Wrap(err, "BodyParser")
	}

	var user User
	err = db.Where("Email = ?", loginData.Email).Find(&user).Error
	if err != nil {
		return errors.Wrap(err, "Get user by email")
	}

	currentUser := SessionCurrentUser{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}
	b, err := json.Marshal(currentUser)
	if err != nil {
		return errors.Wrap(err, "json marshall currentUser")
	}

	session.Set(SessionCurrentUserKey, string(b))
	err = session.Save()
	if err != nil {
		return errors.Wrap(err, "session save()")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "user was logged in",
	})

}

func GetUsers(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	users, err := getUsers(db)
	if err != nil {
		return err
	}

	// TODO: return only requested fields, no created_at,deleted_at,messages etc for all route handlers
	return c.JSON(fiber.Map{
		"Users": users,
	})
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
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	var user User
	userID, err := c.ParamsInt("userID")
	if err != nil {
		return err
	}
	tx := db.Preload("Chats").First(&user, userID)
	if tx.Error != nil {
		return tx.Error
	}

	return c.JSON(fiber.Map{
		"User": user,
	})
}

type FieldError struct {
	Field, Tag, Param string
}

func CreateUser(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	// TODO: create schemas for all models like UserBase, UserCreate, User
	var user User
	err := c.BodyParser(&user)
	if err != nil {
		return err
	}

	validate, ok := c.Locals("validate").(*validator.Validate)
	if !ok {
		log.Fatalf("error getting `validate` from c.Locals()")
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
	tx := db.Create(&user)
	if tx.Error != nil {
		return tx.Error
	}
	return c.JSON(user)
}

type GetChatsResponse struct {
	Chats []Chat
}

func GetChats(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	var chats []Chat
	ch := &Chat{}
	model := db.Model(ch)
	query := model.Preload("Members")
	tx := query.Find(&chats)
	if tx.Error != nil {
		return tx.Error
	}

	return c.JSON(GetChatsResponse{
		Chats: chats,
	})
}

func GetChat(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	var params struct {
		ChatID uint
	}
	err := c.ParamsParser(&params)
	if err != nil {
		return err
	}

	var chat Chat
	err = db.Preload("Members").First(&chat, params.ChatID).Error
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"Chat": chat,
	})
}

type SendMessageRequest struct {
	UserEmail string
	Content   string
}

func SendMessage(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	sessionCurrentUser, err := getLoggedInUser(c)
	if err != nil {
		_, isUnauthorizedUserError := err.(*UnauthorizedUserError)
		if isUnauthorizedUserError {
			return errors.Wrap(err, "isUnauthorizedUserError")
		} else {
			return errors.Wrap(err, "getLoggedInUser()")
		}
	}

	var params struct {
		ChatID int
	}
	err = c.ParamsParser(&params)
	if err != nil {
		return errors.Wrap(err, "ParamsParser")
	}

	var data SendMessageRequest
	err = c.BodyParser(&data)
	if err != nil {
		return errors.Wrap(err, "BodyParser failed")
	}

	validate, ok := c.Locals("validate").(*validator.Validate)
	if !ok {
		log.Fatalf("error getting `validate` from c.Locals()")
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

	userEmail := sessionCurrentUser.Email
	var user User
	tx := db.Where("Email = ?", userEmail).First(&user)
	if tx.Error != nil {
		return tx.Error
	}

	message := Message{
		ChatID:  uint(params.ChatID),
		FromID:  user.ID,
		Content: data.Content,
	}
	tx = db.Create(&message)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "db create message failed")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ID": message.ID,
	})
}

func RootHandler(c *fiber.Ctx) error {
	return c.Redirect("/ui", fiber.StatusPermanentRedirect)
}

func LoginView(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{})
}

type UserAuthSchema struct {
	Email    string
	Password string
}

func PostLoginView(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	var data UserAuthSchema
	err := c.BodyParser(&data)
	if err != nil {
		return err
	}

	store, ok := c.Locals("store").(*session.Store)
	if !ok {
		log.Fatalf("error getting `store` from c.Locals()")
	}

	session, err := store.Get(c)
	if err != nil {
		return err
	}

	var user *User
	tx := db.Where("Email = ?", data.Email).First(&user)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		// NOTE: creating user if it does not exist for v1
		return handleNewUserLogin(c, db, session, data)
	} else if tx.Error != nil {
		return tx.Error
	}

	sessionCurrentUser := SessionCurrentUser{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}
	b, err := json.Marshal(sessionCurrentUser)
	if err != nil {
		return err
	}
	sessionCurrentUserJSON := string(b)
	session.Set(SessionCurrentUserKey, sessionCurrentUserJSON)
	err = session.Save()
	if err != nil {
		return err
	}

	return c.Render("home", fiber.Map{
		"CurrentUser": user,
	})
}

func handleNewUserLogin(c *fiber.Ctx, db *gorm.DB, session *session.Session, data UserAuthSchema) error {
	// password is not used anywhere

	createdUser := User{
		Email:    data.Email,
		Password: data.Password,
	}
	tx := db.Create(&createdUser)
	if tx.Error != nil {
		return tx.Error
	}

	sessionCurrentUser := SessionCurrentUser{
		ID:        createdUser.ID,
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		AvatarURL: createdUser.AvatarURL,
	}
	b, err := json.Marshal(sessionCurrentUser)
	if err != nil {
		return err
	}
	sessionCurrentUserJSON := string(b)
	session.Set(SessionCurrentUserKey, sessionCurrentUserJSON)
	err = session.Save()
	if err != nil {
		return err
	}

	return c.Render("home", fiber.Map{
		"CurrentUser": createdUser,
	})
}

func UploadUserAvatar(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	file, err := c.FormFile("image")
	if err != nil {
		return err
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
	tx := db.Model(&User{}).Where("id = ?", userID).Update("AvatarURL", fmt.Sprintf("/%s", fileName))
	if tx.Error != nil {
		return tx.Error
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}

func JoinChat(c *fiber.Ctx) error {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	var body struct {
		Email string
	}

	err := c.BodyParser(&body)
	if err != nil {
		return errors.Wrap(err, "BodyParser")
	}

	var user User
	tx := db.Where("Email = ?", body.Email).First(&user)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "Filter User by Email")
	}

	var params struct {
		ChatID uint
	}
	err = c.ParamsParser(&params)
	if err != nil {
		return errors.Wrap(err, "ParamsParser")
	}

	var chat Chat
	tx = db.Find(&chat, params.ChatID)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "Find Chat by ID")
	}

	err = db.Model(&chat).Association("Members").Append(&user)
	if err != nil {
		return errors.Wrap(err, "Chat appends member")
	}

	err = db.Save(&user).Error
	if err != nil {
		return errors.Wrap(err, "Save Chat after Members Update")
	}

	return nil
}
