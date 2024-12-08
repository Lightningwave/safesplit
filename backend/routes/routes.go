package routes

import (
	"safesplit/controllers"
	"safesplit/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine,
	loginController *controllers.LoginController,
	createAccountController *controllers.CreateAccountController) {

	api := router.Group("/api")
	{
		// Public routes (no authentication required)
		api.POST("/login", loginController.Login)
		api.POST("/register", createAccountController.CreateAccount)

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User profile routes
			protected.GET("/me", loginController.GetMe)
		}
	}
}
