package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	*gorm.Model

	Name  string `gorm:"uniqueIndex" binding:"required"`
	Email string `gorm:"uniqueIndex" binding:"required"`
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	for _, err := range c.Errors {
		fmt.Println(err)
	}

	if len(c.Errors) > 0 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
	}
}

type GoogleAuthResponse struct {
	credential   string
	g_csrf_token string
}

func main() {

	dsn := "host=0.0.0.0 port=5432 dbname=ginapp sslmode=disable TimeZone=Europe/Kiev"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		panic(err)
	}

	e := gin.Default()

	e.Use(ErrorHandler)

	e.LoadHTMLGlob("./templates/*")

	e.GET("", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	e.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	e.GET("/users", func(c *gin.Context) {
		var users []User
		tx := db.Find(&users)
		if tx.Error != nil {
			c.AbortWithError(http.StatusInternalServerError, tx.Error)
			return
		}
		c.IndentedJSON(http.StatusOK, users)
	})

	e.POST("/user", func(c *gin.Context) {
		var user User
		if err := c.BindJSON(&user); err != nil {
			return
		}
		tx := db.Create(&user)
		if tx.Error != nil {
			c.AbortWithError(http.StatusInternalServerError, tx.Error)
			return
		}
		c.IndentedJSON(http.StatusOK, user)
	})

	e.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"hello": "world",
		})
	})

	e.POST("/login", func(c *gin.Context) {
		bytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		data := string(bytes)
		values, err := url.ParseQuery(data)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		googleAuthResponse := GoogleAuthResponse{
			credential:   values.Get("credential"),
			g_csrf_token: values.Get("g_csrf_token"),
		}
		fmt.Printf("%v\n", googleAuthResponse)

		c.IndentedJSON(http.StatusOK, gin.H{
			"login": "success",
			"data": map[string]string{
				"credential":   googleAuthResponse.credential,
				"g_csrf_token": googleAuthResponse.g_csrf_token,
			},
		})
	})

	e.Run("0.0.0.0:8080")
}

// todo:
// google auth
// apple auth
// cognito auth
// roles permissions
// google pay
// apple pay
// stripe
