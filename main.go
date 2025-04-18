package main

import (
	"project/libraryManagement/config"
	"project/libraryManagement/controllers"
	"project/libraryManagement/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	config.ConnectToDB()
}

func main() {
	r := gin.Default()

	// cors
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "https://library-client-liard.vercel.app"},
        AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Authorization", "token", "Content-Type", "Accept"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    }))
	  

	r.GET("/", controllers.Health)

	// auth routes
	authRoutes := r.Group("/auth")
	authRoutes.POST("/library", controllers.RegisterLibrary)
	authRoutes.POST("/login", controllers.Login)
	authRoutes.POST("/otp/verify", controllers.VerifyUserOTP)

	// owner routes
	ownerRoutes := r.Group("/owner")
	ownerRoutes.Use(middlewares.AuthOwner)
	ownerRoutes.POST("/onboard/admin", controllers.OnboardAdmin)
	ownerRoutes.GET("/admin/list", controllers.RetrieveAdminByLib)

	// admin routes
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middlewares.AuthAdmin)
	adminRoutes.POST("/onboard/reader", controllers.OnboardReader)
	adminRoutes.POST("/create/inventory", controllers.CreateInventory)
	adminRoutes.DELETE("/delete/book/:id", controllers.RemoveBook)
	adminRoutes.POST("/add/book", controllers.AddBook)
	adminRoutes.PATCH("/update/book", controllers.UpdateBook)
	adminRoutes.POST("/issue/approve", controllers.ApproveIssueRequest)
	adminRoutes.POST("/return/approve", controllers.ApproveReturnRequest)
	adminRoutes.POST("/reject/request", controllers.RejectRequest)
	adminRoutes.GET("/reader/list", controllers.RetrieveReaders)

	// reader routes
	readerRoutes := r.Group("/reader")
	readerRoutes.Use(middlewares.AuthReader)
	readerRoutes.POST("/book/search", controllers.SearchBook)
	readerRoutes.POST("/issue/request", controllers.IssueRequest)
	readerRoutes.POST("/return/request", controllers.ReturnRequest)

	// admin + reader routes
	userRoutes := r.Group("/user")
	userRoutes.Use(middlewares.AuthAdminAndReader)
	userRoutes.GET("/issues", controllers.RetrieveRequets)
	userRoutes.GET("/registry", controllers.RetrieveRegistry)

	// public routes
	r.GET("/book/lib/:id", controllers.RetrieveBooksByLib)

	r.Run("0.0.0.0:3001")
}
