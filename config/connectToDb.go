package config

import (
	"fmt"
	"log"
	"os"
	"project/libraryManagement/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error

	envError := godotenv.Load(".env")

	if envError != nil {
		log.Fatalf("Error loading .env file")
		return
	}

	db_url := (os.Getenv("DATABASE_URL"))

	dt := db_url
	DB, err = gorm.Open(postgres.Open(dt), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	DB.AutoMigrate(&models.Library{})
	DB.AutoMigrate(&models.Users{})
	DB.AutoMigrate(&models.BookInventory{})
	DB.AutoMigrate(&models.RequestEvent{})
	DB.AutoMigrate(&models.IssueRegistery{})

	fmt.Println("Connected To Database")
}

