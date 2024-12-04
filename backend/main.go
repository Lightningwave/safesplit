package main

import (
	"log"
	"safesplit/config"
	"safesplit/controllers"
	"safesplit/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize controllers
	loginController := controllers.NewLoginController(db)
	logoutController := controllers.NewLogoutController(db)
	userController := controllers.NewUserController(db)

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

	routes.SetupRoutes(router, loginController, logoutController, userController)

	log.Println("Server starting on :8080")
	router.Run(":8080")
}
