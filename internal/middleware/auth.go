package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		validToken := os.Getenv("AUTH_TOKEN")

		if validToken == "" {
			logrus.Warn("Token Not Found in Environment Variables!")
		}

		if token != validToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "[E] Invalid or missing token!",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
