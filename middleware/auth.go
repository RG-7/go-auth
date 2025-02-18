package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/RG-7/go-auth/helpers"
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Remove Bearer
		authHeader = strings.TrimPrefix(authHeader, "Bearer")

		claims, err := helpers.ValidateToken(authHeader)
		if err != nil {
			log.Printf("Token Validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"err": "invalid token"})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
