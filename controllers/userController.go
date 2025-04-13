package controllers

import (
	"fmt"
	"net/http"
	"project/libraryManagement/config"
	"project/libraryManagement/models"
	"project/libraryManagement/utils"

	"github.com/gin-gonic/gin"
)


type UserOnBoard struct {
	User  string `json:"user"`
}


// onboard Admin
func OnboardAdmin(c *gin.Context) {
	var data UserOnBoard
	var admin models.Users

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if data.User == ""{
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "all fields are required"})
		return
	}

	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	owner, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// check if user already exists
	duplicate, _ := utils.FindUser(data.User)
	if duplicate != nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "user with the same email address already exists"})
		return
	}

	// check if user already onboarded
	flag := config.DB.Preload("Library").Where("lib_id = ? AND role = ?", owner.LibID, "admin").First(&admin)
	if flag.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "only one admin is allowed"})
		return
	}

	user := models.Users{Email: data.User, Role: "admin", LibID: owner.LibID, Library: owner.Library}

	res := config.DB.Create(&user)
	if res.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error onboading user"})
	}

	// message to be sent to the onboarded admin
	message := fmt.Sprintf("Congratulations, you have been onboarded to Our Library Management System as an Admin. You are assigned to %s Library where you will be working as an Admin and onboarding readers. Please login and start managing the inventory.", owner.Library.Name)

	// trigger mail to the admin
	mail := utils.SendMail(user.Email, message, "Admin Onboarding")
	if mail != nil {
		c.IndentedJSON(http.StatusBadGateway, gin.H{"message": "error sending email"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Admin onboarded successfully", "user": user})

}

// Onboard Readers
func OnboardReader(c *gin.Context) {
	var data UserOnBoard

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	owner, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// check if user already exists
	duplicate, _ := utils.FindUser(data.User)
	if duplicate != nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "user with the same email address already exists"})
		return
	}

	user := models.Users{Email: data.User, Role: "reader", LibID: owner.LibID, Library: owner.Library}

	res := config.DB.Create(&user)
	if res.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error onboading user"})
	}

	// message to be sent to the onboarded reader
	message := fmt.Sprintf("Congratulations, you have been onboarded to Our Library Management System as a Reader. You are assigned to %s Library where you can explore and read books. Login to enjoy unlimited reading.", owner.Library.Name)

	// trigger mail to the reader
	mail := utils.SendMail(user.Email, message, "Reader Onboarding")
	if mail != nil {
		c.IndentedJSON(http.StatusBadGateway, gin.H{"message": "error sending email"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Reader onboarded successfully", "user": user})
}

// retrieve user by lib id
func RetrieveAdminByLib(c *gin.Context) {
	var user []models.Users

	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	owner, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}


	res := config.DB.Preload("Library").Where("lib_id = ? AND role = ?", owner.LibID, "admin").Find(&user)
	if res.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "no admin found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "found", "list": user})

}

// retrieve readers
func RetrieveReaders(c *gin.Context){
	var users []models.Users

	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	owner, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	res := config.DB.Preload("Library").Where("lib_id = ? AND role = ?", owner.LibID, "reader").Find(&users)
	if res.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "readers not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "users found", "list": users})
}