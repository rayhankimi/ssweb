package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		token := c.GetHeader("Authorization")

		validToken := os.Getenv("AUTH_TOKEN")
		//validToken := "test"

		if validToken == "" {
			panic("Token Not Found in Environment Variables!")
		}

		if token != validToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "[E] Invalid or missing token!",
			})
		}
		c.Next()
	}
}
