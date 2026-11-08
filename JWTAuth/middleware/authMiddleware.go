package middleware

import (
	"net/http"
	"strconv"
	"strings"

	helper "github.com/SIMCO-TECH/simco/JWTAuth/helpers"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		// Accept both "Bearer <token>" and a bare token
		clientToken := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := helper.ValidateToken(clientToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Set("user_type", claims.UserType)
		c.Set("uid", strconv.FormatUint(uint64(claims.UserId), 10)) // uint -> string
		c.Next()
	}
}
