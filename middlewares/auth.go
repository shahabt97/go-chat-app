package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func AuthHandler(c *gin.Context) {
	sessionRaw, _ := c.Get("session")
	session, _ := sessionRaw.(*sessions.Session)
	if session.Values["username"] == nil {
		c.JSON(403, gin.H{
			"message": "User unAuthorized",
		})
		c.Abort()
		return
	}
	c.Next()
}
