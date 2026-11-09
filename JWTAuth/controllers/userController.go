package controllers

import (
	"errors"
	"net/http"
	"strconv"

	db "github.com/SIMCO-TECH/simco/JWTAuth/database"
	helper "github.com/SIMCO-TECH/simco/JWTAuth/helpers"
	"github.com/SIMCO-TECH/simco/JWTAuth/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var validate = validator.New()

func HashPassword(password string) (string, error) {
	// Implement password hashing logic here (e.g., using bcrypt)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// Return the hashed password and any error encountered
	return string(bytes), err
}

func VerifyPassword(hashedPassword, password string) bool {
	// Implement password verification logic here (e.g., using bcrypt)
	// Return true if the password matches the hashed password, false otherwise
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var count int64
		err := db.DB.Model(&models.User{}).
			Where("email = ? OR phone = ?", user.Email, user.Phone).
			Count(&count).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "This email or phone number already exists"})
			return
		}

		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user.Password = hashedPassword
		user.ID = 0 // Ensure ID is zero to let GORM auto-increment it

		// 1. Save the user (GORM fills user.ID, CreatedAt, UpdatedAt)
		if err := db.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 2. Generate tokens now that the ID exists
		token, refreshToken, err := helper.GenerateAllTokens(user.Email, user.Name, user.UserType, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 3. Store the tokens on the same row
		if err := helper.UpdateAllTokens(token, refreshToken, user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save tokens to the database"})
			return
		}
		user.Token = token
		user.RefreshToken = refreshToken
		if err := db.DB.Model(&user).Updates(map[string]interface{}{
			"token":         token,
			"refresh_token": refreshToken,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save tokens to the database"})
			return
		}

		user.Password = "" // Clear the password before sending the response
		c.JSON(http.StatusCreated, user)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var foundUser models.User
		if err := db.DB.Where("email = ?", input.Email).First(&foundUser).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		if !VerifyPassword(foundUser.Password, input.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		token, refreshToken, err := helper.GenerateAllTokens(foundUser.Email, foundUser.Name, foundUser.UserType, foundUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate tokens"})
			return
		}
		if err := helper.UpdateAllTokens(token, refreshToken, foundUser.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save tokens"})
			return
		}
		foundUser.Token = token
		foundUser.RefreshToken = refreshToken
		foundUser.Password = "" // Clear the password before sending the response

		c.JSON(http.StatusOK, foundUser)

	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		var total int64
		db.DB.Model(&models.User{}).Count(&total)

		var users []models.User
		if err := db.DB.Select("id", "name", "email", "phone", "user_type", "created_at", "updated_at").
			Order("id").Limit(limit).Offset((page - 1) * limit).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"total_count": total,
			"page":        page,
			"users":       users,
		})
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUserId(c, userId); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var user models.User
		err := db.DB.Where("id = ?", userId).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user.Password = "" // Clear the password before sending the response
		c.JSON(http.StatusOK, user)
	}
}
