package middlewares

import (
	"first/session"

	"github.com/gin-gonic/gin"
)

func SessiosnMiddleware(c *gin.Context) {

	session, _ := session.Store.Get(c.Request, "log-session")

	c.Set("session", session)
	c.Next()

}
