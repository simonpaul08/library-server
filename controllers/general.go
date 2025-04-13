package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type EmailData struct {
	Email string `json:"email"`
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"server is up and running"})
}
