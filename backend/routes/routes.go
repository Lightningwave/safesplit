package routes

import (
	"safesplit/controllers"
	"safesplit/controllers/EndUser"
	"safesplit/controllers/SuperAdmin"
	"safesplit/controllers/SysAdmin"
	"safesplit/middleware"
	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouteHandlers struct {
	LoginController         *controllers.LoginController
	CreateAccountController *controllers.CreateAccountController
	SuperAdminHandlers      *SuperAdminHandlers
	SysAdminHandlers        *SysAdminHandlers
	EndUserHandlers         *EndUserHandlers
}

type EndUserHandlers struct {
	UploadFileController   *EndUser.UploadFileController
	ViewFilesController    *EndUser.ViewFilesController
	DownloadFileController *EndUser.DownloadFileController
	DeleteFileController   *EndUser.DeleteFileController
	ArchiveFileController  *EndUser.ArchiveFileController
	ShareFileController    *EndUser.ShareFileController
}

type SuperAdminHandlers struct {
	CreateSysAdminController *SuperAdmin.CreateSysAdminController
	ViewSysAdminController   *SuperAdmin.ViewSysAdminController
	DeleteSysAdminController *SuperAdmin.DeleteSysAdminController
	SystemLogsController     *SuperAdmin.SystemLogsController
}

type SysAdminHandlers struct {
	UpdateAccountController          *SysAdmin.UpdateAccountController
	DeleteUserAccountController      *SysAdmin.DeleteUserAccountController
	ViewUserAccountController        *SysAdmin.ViewUserAccountController
	ViewDeletedUserAccountController *SysAdmin.ViewDeletedUserAccountController
	ViewUserStorageController        *SysAdmin.ViewUserStorageController
	ViewUserAccountDetailsController *SysAdmin.ViewUserAccountDetailsController
}

func NewRouteHandlers(
	db *gorm.DB,
	userModel *models.UserModel,
	activityLogModel *models.ActivityLogModel,
	fileModel *models.FileModel,
	fileShareModel *models.FileShareModel,
	keyFragmentModel *models.KeyFragmentModel,
	encryptionService *services.EncryptionService,
	shamirService *services.ShamirService,
) *RouteHandlers {
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
			UpdateAccountController:          SysAdmin.NewUpdateAccountController(userModel),
			DeleteUserAccountController:      SysAdmin.NewDeleteUserAccountController(userModel),
			ViewUserAccountController:        SysAdmin.NewViewUserAccountController(userModel),
			ViewDeletedUserAccountController: SysAdmin.NewViewDeletedUserAccountController(userModel),
			ViewUserStorageController:        SysAdmin.NewViewUserStorageController(userModel),
			ViewUserAccountDetailsController: SysAdmin.NewViewUserAccountDetailsController(userModel),
		},
		EndUserHandlers: &EndUserHandlers{
			UploadFileController:   EndUser.NewFileController(fileModel, activityLogModel, encryptionService, shamirService, keyFragmentModel),
			ViewFilesController:    EndUser.NewViewFilesController(fileModel),
			DownloadFileController: EndUser.NewDownloadFileController(fileModel, keyFragmentModel, encryptionService, activityLogModel),
			DeleteFileController:   EndUser.NewDeleteFileController(fileModel),
			ArchiveFileController:  EndUser.NewArchiveFileController(fileModel),
			ShareFileController:    EndUser.NewShareFileController(fileModel, fileShareModel, keyFragmentModel, encryptionService, activityLogModel),
		},
	}
}

func SetupRoutes(router *gin.Engine, handlers *RouteHandlers, userModel *models.UserModel) {
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
	api.POST("/files/share/:shareLink", handlers.EndUserHandlers.ShareFileController.AccessShare)
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func setupProtectedRoutes(protected *gin.RouterGroup, handlers *RouteHandlers) {
	protected.GET("/me", handlers.LoginController.GetMe)

	// End User routes should be first as they're most commonly accessed
	setupEndUserRoutes(protected, handlers.EndUserHandlers)

	// Admin routes with their respective middleware
	superAdmin := protected.Group("/admin")
	superAdmin.Use(middleware.SuperAdminMiddleware())
	setupSuperAdminRoutes(superAdmin, handlers.SuperAdminHandlers)

	sysAdmin := protected.Group("/system")
	sysAdmin.Use(middleware.SysAdminMiddleware())
	setupSysAdminRoutes(sysAdmin, handlers.SysAdminHandlers)
}

func setupEndUserRoutes(protected *gin.RouterGroup, handlers *EndUserHandlers) {
	files := protected.Group("/files")
	{
		files.GET("", handlers.ViewFilesController.ListUserFiles)
		files.GET("/:id/download", handlers.DownloadFileController.Download)
		files.POST("/upload", handlers.UploadFileController.Upload)
		files.DELETE("/:id", handlers.DeleteFileController.Delete)
		files.PUT("/:id/archive", handlers.ArchiveFileController.Archive)
		files.POST("/:id/share", handlers.ShareFileController.CreateShare)
		files.GET("/share/:shareLink", handlers.ShareFileController.AccessShare)
	}
}

func setupSuperAdminRoutes(superAdmin *gin.RouterGroup, handlers *SuperAdminHandlers) {
	superAdmin.POST("/create-sysadmin", handlers.CreateSysAdminController.CreateSysAdmin)
	superAdmin.GET("/sysadmins", handlers.ViewSysAdminController.ListSysAdmins)
	superAdmin.DELETE("/sysadmins/:id", handlers.DeleteSysAdminController.DeleteSysAdmin)
	superAdmin.GET("/system-logs", handlers.SystemLogsController.GetSystemLogs)
}

func setupSysAdminRoutes(sysAdmin *gin.RouterGroup, handlers *SysAdminHandlers) {
	userRoutes := sysAdmin.Group("/users")
	{
		userRoutes.GET("", handlers.ViewUserAccountController.ListUsers)
		userRoutes.GET("/:id", handlers.ViewUserAccountDetailsController.GetUserAccountDetails)
		userRoutes.PUT("/:id", handlers.UpdateAccountController.UpdateAccount)
		userRoutes.DELETE("/:id", handlers.DeleteUserAccountController.DeleteUser)

		userRoutes.GET("/deleted", handlers.ViewDeletedUserAccountController.GetDeletedUsers)
		userRoutes.POST("/deleted/:id/restore", handlers.ViewDeletedUserAccountController.RestoreUser)
	}

	sysAdmin.GET("/storage/stats", handlers.ViewUserStorageController.GetStorageStats)
}
