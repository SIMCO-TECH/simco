package main

import (
	"log"
	"os"
	"simco/JWTAuth/database"
	"simco/JWTAuth/models"
	routes "simco/JWTAuth/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	database.ConnectDB()
	if err := database.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from API-1"})
	})

	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from API-2"})
	})

	router.Run(":" + port)
}
