package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"project/libraryManagement/controllers"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	return router
}

func TestHealth(t *testing.T) {
	mockResponse := `{"message":"server is up and running"}`

	r := SetupRouter()
	r.GET("/", controllers.Health)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	responseData, _ := io.ReadAll(w.Body)
	assert.Equal(t, mockResponse, string(responseData))
	assert.Equal(t, http.StatusOK, w.Code)
}
