package routes

import (
	"safesplit/controllers"
	"safesplit/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, loginController *controllers.LoginController,
	logoutController *controllers.LogoutController, userController *controllers.UserController) {

	api := router.Group("/api")
	{
		// Public routes
		api.POST("/login", loginController.Login)
		api.POST("/register", userController.Create)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/logout", logoutController.Logout)
			protected.GET("/me", userController.GetMe)

			// Premium User routes
			premium := protected.Group("/premium")
			premium.Use(middleware.RequirePremiumUser())
			{
				// Add premium routes here
			}

			// System Admin routes
			sysAdmin := protected.Group("/admin")
			sysAdmin.Use(middleware.RequireSysAdmin())
			{
				// Add system admin routes here
			}

			// Super Admin routes
			superAdmin := protected.Group("/super")
			superAdmin.Use(middleware.RequireSuperAdmin())
			{
				// Add super admin routes here
			}
		}
	}
}
