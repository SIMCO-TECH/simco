package database

import (
	"fmt"
	"time"
	"context"
	"log"
	"os"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PostgresDb := os.Getenv("POSTGRES_URL")
}
