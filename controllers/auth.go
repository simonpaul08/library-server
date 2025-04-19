package controllers

import (
	"net/http"
	"project/libraryManagement/utils"

	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Email string `json:"email"`
}

type VerifyOTPData struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// Login
func Login(c *gin.Context) {
	var data LoginData

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if user exists
	user, _ := utils.FindUser(data.Email)
	if user == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user doesn't exists"})
		return
	}

	// check if user is admin, owner or a reader
	if user.Role == "owner" || user.Role == "admin" {

		// send otp for verification
		result := utils.SendOTP(data.Email)
		if result != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error sending otp"})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "otp sent successfully"})
	} else {

		// send otp for verification
		result := utils.SendOTP(data.Email)
		if result != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error sending otp"})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "otp sent successfully"})
	}

}

// verify OTP
func VerifyUserOTP(c *gin.Context) {
	var data VerifyOTPData

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userError := utils.FindUser(data.Email)
	if userError != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user doesn't exists"})
		return
	}

	// verify otp
	user, e := utils.VerifyOTP(data.Email, data.OTP)
	if e != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": e.Error()})
		return
	}

	// gererate token
	token, err := utils.GenerateToken(user)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"messsage": "error generating token", "error": err})
		return
	}

	c.SetCookie("token", token, 3600, "/", "localhost", false, true)
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Otp verified", "user": user, "token": token})
}

// demo login
func DemoLogin(c *gin.Context) {
	var data LoginData

	err := c.ShouldBind((&data))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	user, _ := utils.FindUser(data.Email)

	if user == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user doesn't exists"})
		return
	}

	// gererate token
	token, err := utils.GenerateToken(user)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"messsage": "error generating token", "error": err})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"user": user, "token": token})

}
