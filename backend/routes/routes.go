package routes

import (
	"safesplit/controllers"
	"safesplit/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, userController *controllers.UserController, authController *controllers.AuthController) {
	api := router.Group("/api")
	{
		// Public routes
		api.POST("/login", authController.Login)
		api.POST("/users", userController.Create) // signup

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Add protected routes here
			protected.GET("/users/me", userController.GetMe)
		}
	}
}
