package controllers

import (
	"fmt"
	"os"
	"time"

	// "path"
	// "path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func HomePageHandler(c *gin.Context) {
	sessionRaw, _ := c.Get("session")
	session, _ := sessionRaw.(*sessions.Session)
	if session.Values["username"] != nil {
		c.Redirect(302, "/chat/public")
		return
	}
	c.File("views/login.html")
}

func FileHandler(c *gin.Context) {
	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(400, gin.H{
			"message": "something went wrong",
		})
		return
	}

	filename := file.Filename
	wd, _ := os.Getwd()
	fmt.Println("p: ", wd)
	timeCreated := time.Now()
	filePathInWD := fmt.Sprintf("/public/images/%s_%s", timeCreated.String(), filename)
	filepathInServer := fmt.Sprintf("%s%s", wd, filePathInWD)
	// extention := filepath.Ext(filename)

	c.SaveUploadedFile(file, filepathInServer)

	fmt.Println("filename is: ", filepathInServer)
	c.JSON(200,filePathInWD)
}
