package main

import (
	"github.com/SIMCO-TECH/simco/JWTAuth/database"
	"github.com/SIMCO-TECH/simco/JWTAuth/models"
	routes "github.com/SIMCO-TECH/simco/JWTAuth/routes"
	"log"
	"os"

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

	routes.AuthRoutes(router) // signup and login stay public
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from API-1"})
	})

	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from API-2"})
	})

	router.Run(":" + port)
}
