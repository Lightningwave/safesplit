package routes

import (
	"safesplit/controllers"
	"safesplit/controllers/SuperAdmin"
	"safesplit/middleware"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type RouteHandlers struct {
	LoginController         *controllers.LoginController
	CreateAccountController *controllers.CreateAccountController
	SuperAdminHandlers      *SuperAdminHandlers
}

type SuperAdminHandlers struct {
	CreateSysAdminController *SuperAdmin.CreateSysAdminController
}

func NewRouteHandlers(userModel *models.UserModel) *RouteHandlers {
	return &RouteHandlers{
		LoginController:         controllers.NewLoginController(userModel),
		CreateAccountController: controllers.NewCreateAccountController(userModel),
		SuperAdminHandlers: &SuperAdminHandlers{
			CreateSysAdminController: SuperAdmin.NewCreateSysAdminController(userModel),
		},
	}
}

func SetupRoutes(
	router *gin.Engine,
	handlers *RouteHandlers,
	userModel *models.UserModel,
) {
	// API group
	api := router.Group("/api")
	{
		// Public routes
		setupPublicRoutes(api, handlers)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(userModel))
		setupProtectedRoutes(protected, handlers)
	}
}

func setupPublicRoutes(api *gin.RouterGroup, handlers *RouteHandlers) {
	api.POST("/login", handlers.LoginController.Login)
	api.POST("/register", handlers.CreateAccountController.CreateAccount)

	// Health check endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func setupProtectedRoutes(protected *gin.RouterGroup, handlers *RouteHandlers) {
	// User profile routes
	protected.GET("/me", handlers.LoginController.GetMe)

	// Super Admin routes
	superAdmin := protected.Group("/admin")
	superAdmin.Use(middleware.SuperAdminMiddleware())
	setupSuperAdminRoutes(superAdmin, handlers.SuperAdminHandlers)
}

func setupSuperAdminRoutes(superAdmin *gin.RouterGroup, handlers *SuperAdminHandlers) {
	// System Administrator management
	superAdmin.POST("/create-sysadmin", handlers.CreateSysAdminController.CreateSysAdmin)
}
