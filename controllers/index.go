package controllers

import (
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
