package controllers

import "github.com/gin-gonic/gin"

func PublicChatHandler(c *gin.Context) {
	c.File("views/public-chat.html")
}


