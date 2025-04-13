package middlewares

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// authentication middleware
func AuthOwner(c *gin.Context) {
	clientToken := c.Request.Header.Get("Authorization")

	if clientToken == "" {

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "no authorization token provided"})

		c.Abort()

		return

	}
	secret := []byte(os.Getenv("SECRET"))

	token, err := jwt.Parse(clientToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error validating token", "error": err.Error()})
		c.Abort()
		return
	}

	if !token.Valid {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "invalid token"})
		c.Abort()
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	fmt.Printf("claims %v", claims["role"])

	if claims["role"] != "owner" {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, only owner can access the route"})
		c.Abort()
		return
	}

	c.Set("email", claims["email"])

	c.Next()
}

// validate Admin
func AuthAdmin(c *gin.Context) {
	clientToken := c.Request.Header.Get("Authorization")
	if clientToken == "" {

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "no authorization token provided"})

		c.Abort()

		return

	}
	secret := []byte(os.Getenv("SECRET"))

	token, err := jwt.Parse(clientToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error validating token", "error": err.Error()})
		c.Abort()
		return
	}

	if !token.Valid {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "invalid token"})
		c.Abort()
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	fmt.Printf("claims %v", claims["role"])

	if claims["role"] != "admin" {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, only admin can access the route"})
		c.Abort()
		return
	}

	c.Set("email", claims["email"])

	c.Next()
}

// validate reader
func AuthReader(c *gin.Context) {
	clientToken := c.Request.Header.Get("Authorization")
	if clientToken == "" {

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "no authorization token provided"})

		c.Abort()

		return

	}
	secret := []byte(os.Getenv("SECRET"))

	token, err := jwt.Parse(clientToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error validating token", "error": err.Error()})
		c.Abort()
		return
	}

	if !token.Valid {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "invalid token"})
		c.Abort()
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	fmt.Printf("claims %v", claims["role"])

	if claims["role"] != "reader" {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, only reader can access the route"})
		c.Abort()
		return
	}

	c.Set("email", claims["email"])

	c.Next()
}

// validate both reader and admin
func AuthAdminAndReader(c *gin.Context) {
	clientToken := c.Request.Header.Get("Authorization")
	if clientToken == "" {

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "no authorization token provided"})

		c.Abort()

		return

	}
	secret := []byte(os.Getenv("SECRET"))

	token, err := jwt.Parse(clientToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error validating token", "error": err.Error()})
		c.Abort()
		return
	}

	if !token.Valid {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "invalid token"})
		c.Abort()
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	fmt.Printf("claims %v", claims["role"])

	if claims["role"] == "owner" {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, only admin and reader can access the route"})
		c.Abort()
		return
	}

	c.Set("email", claims["email"])

	c.Next()
}
