package config

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"project/libraryManagement/models"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error

	dt := "postgres://postgres:postgres@localhost:5432/library?connect_timeout=10"
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

// postgres://ohfvvnou:sJ1qyOSbStI57eDBbC8dOVXNQI1OWwZj@tyke.db.elephantsql.com/ohfvvnou
