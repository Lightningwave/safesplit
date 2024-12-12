package routes

import (
	"safesplit/controllers"
	"safesplit/controllers/SuperAdmin"
	"safesplit/controllers/SysAdmin"
	"safesplit/middleware"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type RouteHandlers struct {
	LoginController         *controllers.LoginController
	CreateAccountController *controllers.CreateAccountController
	SuperAdminHandlers      *SuperAdminHandlers
	SysAdminHandlers        *SysAdminHandlers
}

type SuperAdminHandlers struct {
	CreateSysAdminController *SuperAdmin.CreateSysAdminController
	ViewSysAdminController   *SuperAdmin.ViewSysAdminController
	DeleteSysAdminController *SuperAdmin.DeleteSysAdminController
	SystemLogsController     *SuperAdmin.SystemLogsController
}

type SysAdminHandlers struct {
	UpdateUserAccessController       *SysAdmin.UpdateUserAccessController
	DeleteUserAccountController      *SysAdmin.DeleteUserAccountController
	ViewUserAccountController        *SysAdmin.ViewUserAccountController
	ViewDeletedUserAccountController *SysAdmin.ViewDeletedUserAccountController
	ViewUserStorageController        *SysAdmin.ViewUserStorageController
	ViewUserSubscriptionController   *SysAdmin.ViewUserSubscriptionController
}

func NewRouteHandlers(userModel *models.UserModel, activityLogModel *models.ActivityLogModel) *RouteHandlers {
	return &RouteHandlers{
		LoginController:         controllers.NewLoginController(userModel),
		CreateAccountController: controllers.NewCreateAccountController(userModel),
		SuperAdminHandlers: &SuperAdminHandlers{
			CreateSysAdminController: SuperAdmin.NewCreateSysAdminController(userModel),
			ViewSysAdminController:   SuperAdmin.NewViewSysAdminController(userModel),
			DeleteSysAdminController: SuperAdmin.NewDeleteSysAdminController(userModel),
			SystemLogsController:     SuperAdmin.NewSystemLogsController(activityLogModel),
		},
		SysAdminHandlers: &SysAdminHandlers{
			UpdateUserAccessController:       SysAdmin.NewUpdateUserAccessController(userModel),
			DeleteUserAccountController:      SysAdmin.NewDeleteUserAccountController(userModel),
			ViewUserAccountController:        SysAdmin.NewViewUserAccountController(userModel),
			ViewDeletedUserAccountController: SysAdmin.NewViewDeletedUserAccountController(userModel), // Add this
			ViewUserStorageController:        SysAdmin.NewViewUserStorageController(userModel),
			ViewUserSubscriptionController:   SysAdmin.NewViewUserSubscriptionController(userModel),
		},
	}
}

func SetupRoutes(
	router *gin.Engine,
	handlers *RouteHandlers,
	userModel *models.UserModel,
) {
	api := router.Group("/api")
	{
		setupPublicRoutes(api, handlers)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(userModel))
		setupProtectedRoutes(protected, handlers)
	}
}

func setupPublicRoutes(api *gin.RouterGroup, handlers *RouteHandlers) {
	api.POST("/login", handlers.LoginController.Login)
	api.POST("/register", handlers.CreateAccountController.CreateAccount)
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func setupProtectedRoutes(protected *gin.RouterGroup, handlers *RouteHandlers) {
	protected.GET("/me", handlers.LoginController.GetMe)

	superAdmin := protected.Group("/admin")
	superAdmin.Use(middleware.SuperAdminMiddleware())
	setupSuperAdminRoutes(superAdmin, handlers.SuperAdminHandlers)

	sysAdmin := protected.Group("/system")
	sysAdmin.Use(middleware.SysAdminMiddleware())
	setupSysAdminRoutes(sysAdmin, handlers.SysAdminHandlers)
}

func setupSuperAdminRoutes(superAdmin *gin.RouterGroup, handlers *SuperAdminHandlers) {
	superAdmin.POST("/create-sysadmin", handlers.CreateSysAdminController.CreateSysAdmin)
	superAdmin.GET("/sysadmins", handlers.ViewSysAdminController.ListSysAdmins)
	superAdmin.DELETE("/sysadmins/:id", handlers.DeleteSysAdminController.DeleteSysAdmin)
	superAdmin.GET("/system-logs", handlers.SystemLogsController.GetSystemLogs)
}

func setupSysAdminRoutes(sysAdmin *gin.RouterGroup, handlers *SysAdminHandlers) {
	// User management routes
	userRoutes := sysAdmin.Group("/users")
	{
		userRoutes.GET("", handlers.ViewUserAccountController.ListUsers)
		userRoutes.DELETE("/:id", handlers.DeleteUserAccountController.DeleteUser)
		userRoutes.PUT("/:id/access", handlers.UpdateUserAccessController.UpdateUserAccess)

		// Deleted users routes
		userRoutes.GET("/deleted", handlers.ViewDeletedUserAccountController.GetDeletedUsers)
		userRoutes.POST("/deleted/:id/restore", handlers.ViewDeletedUserAccountController.RestoreUser)
	}

	// Stats routes
	sysAdmin.GET("/storage/stats", handlers.ViewUserStorageController.GetStorageStats)
	sysAdmin.GET("/subscription/stats", handlers.ViewUserSubscriptionController.GetSubscriptionStats)
}
