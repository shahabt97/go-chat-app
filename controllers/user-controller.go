package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"go-chat-app/database"

	"go-chat-app/elasticsearch"
	"go-chat-app/session"
	"go-chat-app/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// register
func RegisterHandler(c *gin.Context) {

	registerData := &struct {
		Username string `form:"username"`
		Email    string `form:"email"`
		Password string `form:"password"`
	}{}

	if err := c.Bind(registerData); err != nil {
		println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	password := utils.HashPassword(registerData.Password)
	if password == nil {
		fmt.Println("there is an error in bycrypting password")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}

	result, err := database.Users.InsertOne(ctx, bson.D{
		{Key: "username", Value: registerData.Username},
		{Key: "email", Value: registerData.Email},
		{Key: "password", Value: password},
		{Key: "joinedAt", Value: time.Now()},
	})
	if err != nil {
		fmt.Println("error in inserting user data in database: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}

	userDataForElastic := &elasticsearch.UserIndex{
		Id:       result.InsertedID.(primitive.ObjectID).Hex(),
		Username: registerData.Username,
		Email:    registerData.Email,
	}

	go func(data *elasticsearch.UserIndex) {
		jsonData, err := json.Marshal(userDataForElastic)
		if err != nil {
			fmt.Println("Error in Marshaling user data for elastic: ", err)
			return
		}
		dataReader := bytes.NewReader(jsonData)
		errOfEs := elasticsearch.Client.CreateDoc("users", dataReader)
		if errOfEs != nil {
			fmt.Println("Error in creating user in elastic: ", errOfEs)
			return
		}
	}(userDataForElastic)
	c.Redirect(http.StatusFound, "http://localhost:8080/user/login")
}

// login
func LoginHandler(c *gin.Context) {

	loginData := &struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}{}
	if err := c.Bind(loginData); err != nil {
		println(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := database.Users.FindOne(ctx, bson.M{
		"username": loginData.Username,
	})

	if result.Err() != nil {
		fmt.Println("error in finding user data in database: ", result.Err().Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": result.Err().Error(),
		})
		return
	}

	// data of the user that we want
	userData := &struct {
		ID       string `bson:"_id"`
		Username string `bson:"username"`
		Password string `bson:"password"`
	}{}

	err := result.Decode(userData)

	if err != nil {
		fmt.Println("error in decoding login data", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err2 := utils.CompareHash(userData.Password, loginData.Password); err2 != nil {
		fmt.Println("password is incorrect: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "password is incorrect",
		})
		return
	}

	session, _ := session.Store.Get(c.Request, "log-session")
	session.Options.MaxAge = int(12 * 24 * time.Hour / time.Second)
	session.Options.HttpOnly = true
	session.Values["username"] = userData.Username
	session.Values["id"] = userData.ID

	err_saving_session := session.Save(c.Request, c.Writer)
	if err_saving_session != nil {
		fmt.Println("err: ", err_saving_session)
		return
	}

	cookie := &http.Cookie{
		Name:  "username",
		Value: userData.Username,
		Path:  "/",
		// HttpOnly: true,
		Expires: time.Now().Add(12 * 24 * time.Hour),
	}
	http.SetCookie(c.Writer, cookie)

	c.Redirect(301, "/chat/public")

}

func GetLoginPage(c *gin.Context) {
	c.File("views/login.html")
}

func GetUserInfoFromSession(c *gin.Context) {

	sessionRaw, exists := c.Get("session")
	if !exists {
		c.JSON(403, nil)
		return
	}
	session, ok := sessionRaw.(*sessions.Session)
	if !ok {
		c.JSON(500, nil)
		return
	}
	c.JSON(200, gin.H{
		"username": session.Values["username"],
		"id":       session.Values["id"],
	})
}

func LogoutHandler(c *gin.Context) {

	sessionRaw, _ := c.Get("session")
	session, _ := sessionRaw.(*sessions.Session)

	cookie := &http.Cookie{
		Name:    "username",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(c.Writer, cookie)

	session.Options.MaxAge = -1
	session.Save(c.Request, c.Writer)

	c.Redirect(302, "/")
}

func RegisterPage(c *gin.Context) {
	c.File("views/register.html")
}

func SearchInUsers(c *gin.Context) {

	query := c.Query("q")

	res, err := elasticsearch.Client.SearchUser(query, "users")
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Server error",
		})
		return
	}
	c.JSON(200, res)
}
