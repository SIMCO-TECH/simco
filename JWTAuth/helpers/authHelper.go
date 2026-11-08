package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(c *gin.Context, role string)  error {
	if c.GetString("user_type") != role {
		return errors.New("Unauthorized to access this resource")
	}
	return nil
}

func MatchUserTypeToUserId(c *gin.Context, userId string)  error {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	
	// A normal USER may only access their own record. ADMIN can access any.
	if userType == "USER" && uid != userId {
		return errors.New("Unauthorized to access this resource")
	}
	return nil
}

