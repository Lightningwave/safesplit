package main

import (
	"log"
	"safesplit/config"
	"safesplit/controllers"
	"safesplit/middleware"
	"safesplit/models"
	"safesplit/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize models
	userModel := models.NewUserModel(db)

	// Initialize controllers
	loginController := controllers.NewLoginController(userModel)
	logoutController := controllers.NewLogoutController(userModel)
	createAccountController := controllers.NewCreateAccountController(userModel)

	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Auth middleware
	router.Use(middleware.AuthMiddleware())

	routes.SetupRoutes(router, loginController, logoutController, createAccountController)

	log.Println("Server starting on :8080")
	router.Run(":8080")
}
