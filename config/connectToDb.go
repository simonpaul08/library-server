package config

import (
	"fmt"
	"os"
	"project/libraryManagement/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error

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

