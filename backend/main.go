package main

import (
	"log"
	"safesplit/config"
	"safesplit/controllers"
	"safesplit/models"
	"safesplit/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize models
	userModel := models.NewUserModel(db)

	// Initialize controllers
	loginController := controllers.NewLoginController(userModel)
	createAccountController := controllers.NewCreateAccountController(userModel)

	// Setup Gin router
	router := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"} // Add your frontend URL
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	router.Use(cors.New(corsConfig))

	// Setup routes
	routes.SetupRoutes(router, loginController, createAccountController)

	// Start server
	port := ":8080"
	log.Printf("Server starting on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
