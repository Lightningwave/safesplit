package routes

import (
	"safesplit/controllers"
	"safesplit/controllers/EndUser"
	"safesplit/controllers/PremiumUser"
	"safesplit/controllers/SuperAdmin"
	"safesplit/controllers/SysAdmin"
	"safesplit/middleware"
	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouteHandlers struct {
	LoginController           *controllers.LoginController
	SuperAdminLoginController *SuperAdmin.LoginController
	CreateAccountController   *controllers.CreateAccountController
	SuperAdminHandlers        *SuperAdminHandlers
	SysAdminHandlers          *SysAdminHandlers
	EndUserHandlers           *EndUserHandlers
	PremiumUserHandlers       *PremiumUserHandlers
}

type EndUserHandlers struct {
	UploadFileController    *EndUser.UploadFileController
	ViewFilesController     *EndUser.ViewFilesController
	DownloadFileController  *EndUser.DownloadFileController
	DeleteFileController    *EndUser.DeleteFileController
	ArchiveFileController   *EndUser.ArchiveFileController
	ShareFileController     *EndUser.ShareFileController
	CreateFolderController  *EndUser.CreateFolderController
	ViewFolderController    *EndUser.ViewFolderController
	DeleteFolderController  *EndUser.DeleteFolderController
	PasswordResetController *EndUser.PasswordResetController
	ViewStorageController   *EndUser.ViewStorageController
}
type PremiumUserHandlers struct {
	FragmentController          *PremiumUser.FragmentController
	FileRecoveryController      *PremiumUser.FileRecoveryController
	AdvancedShareFileController *PremiumUser.ShareFileController
}

type SuperAdminHandlers struct {
	LoginController          *SuperAdmin.LoginController
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
	billingModel *models.BillingModel,
	passwordHistoryModel *models.PasswordHistoryModel,
	activityLogModel *models.ActivityLogModel,
	fileModel *models.FileModel,
	folderModel *models.FolderModel,
	fileShareModel *models.FileShareModel,
	keyFragmentModel *models.KeyFragmentModel,
	encryptionService *services.EncryptionService,
	shamirService *services.ShamirService,
	compressionService *services.CompressionService,
) *RouteHandlers {
	superAdminLoginController := SuperAdmin.NewLoginController(userModel)
	return &RouteHandlers{
		LoginController:           controllers.NewLoginController(userModel, billingModel),
		SuperAdminLoginController: superAdminLoginController,
		CreateAccountController:   controllers.NewCreateAccountController(userModel, passwordHistoryModel),
		SuperAdminHandlers: &SuperAdminHandlers{
			LoginController:          superAdminLoginController,
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
			ViewUserAccountDetailsController: SysAdmin.NewViewUserAccountDetailsController(userModel, billingModel),
		},
		EndUserHandlers: &EndUserHandlers{
			UploadFileController:    EndUser.NewFileController(fileModel, userModel, activityLogModel, encryptionService, shamirService, keyFragmentModel, compressionService, folderModel),
			ViewFilesController:     EndUser.NewViewFilesController(fileModel, folderModel),
			DownloadFileController:  EndUser.NewDownloadFileController(fileModel, keyFragmentModel, encryptionService, activityLogModel, compressionService),
			DeleteFileController:    EndUser.NewDeleteFileController(fileModel),
			ArchiveFileController:   EndUser.NewArchiveFileController(fileModel),
			ShareFileController:     EndUser.NewShareFileController(fileModel, fileShareModel, keyFragmentModel, encryptionService, activityLogModel),
			CreateFolderController:  EndUser.NewCreateFolderController(folderModel, activityLogModel),
			ViewFolderController:    EndUser.NewViewFolderController(folderModel, fileModel),
			DeleteFolderController:  EndUser.NewDeleteFolderController(folderModel, activityLogModel),
			PasswordResetController: EndUser.NewPasswordResetController(userModel, passwordHistoryModel),
			ViewStorageController:   EndUser.NewViewStorageController(fileModel, userModel),
		},
		PremiumUserHandlers: &PremiumUserHandlers{
			FragmentController:          PremiumUser.NewFragmentController(keyFragmentModel, fileModel),
			FileRecoveryController:      PremiumUser.NewFileRecoveryController(fileModel),
			AdvancedShareFileController: PremiumUser.NewShareFileController(fileModel, fileShareModel, keyFragmentModel, encryptionService, activityLogModel),
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
	api.POST("/super-login", handlers.SuperAdminLoginController.Login)
	api.POST("/register", handlers.CreateAccountController.CreateAccount)
	api.POST("/files/share/:shareLink", handlers.EndUserHandlers.ShareFileController.AccessShare)
	api.POST("/premium/shares/:shareLink", handlers.PremiumUserHandlers.AdvancedShareFileController.AccessShare)
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func setupProtectedRoutes(protected *gin.RouterGroup, handlers *RouteHandlers) {
	protected.GET("/me", handlers.LoginController.GetMe)

	// End User routes should be first as they're most commonly accessed
	setupEndUserRoutes(protected, handlers.EndUserHandlers)

	// Premium User routes
	premium := protected.Group("/premium")
	premium.Use(middleware.PremiumUserMiddleware())
	setupPremiumUserRoutes(premium, handlers.PremiumUserHandlers)

	// Admin routes with their respective middleware
	superAdmin := protected.Group("/admin")
	superAdmin.Use(middleware.SuperAdminMiddleware())
	setupSuperAdminRoutes(superAdmin, handlers.SuperAdminHandlers)

	sysAdmin := protected.Group("/system")
	sysAdmin.Use(middleware.SysAdminMiddleware())
	setupSysAdminRoutes(sysAdmin, handlers.SysAdminHandlers)
}

func setupEndUserRoutes(protected *gin.RouterGroup, handlers *EndUserHandlers) {
	protected.PUT("/reset-password", handlers.PasswordResetController.ResetPassword)
	// Existing files routes
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

	folders := protected.Group("/folders")
	{
		folders.GET("", handlers.ViewFolderController.ListFolders)           // Get root folders
		folders.GET("/:id", handlers.ViewFolderController.GetFolderContents) // Get folder contents
		folders.POST("", handlers.CreateFolderController.Create)             // Create new folder
		folders.DELETE("/:id", handlers.DeleteFolderController.Delete)       // Delete folder
	}
	storage := protected.Group("/storage")
	{
		storage.GET("/info", handlers.ViewStorageController.GetStorageInfo)
	}
}
func setupPremiumUserRoutes(premium *gin.RouterGroup, handlers *PremiumUserHandlers) {
	// Fragment management routes
	fragments := premium.Group("/fragments")
	{
		fragments.GET("/files/:fileId", handlers.FragmentController.GetUserFragments)
	}
	recovery := premium.Group("/recovery")
	{
		recovery.GET("/files", handlers.FileRecoveryController.ListRecoverableFiles)
		recovery.POST("/files/:fileId", handlers.FileRecoveryController.RecoverFile)
	}
	shares := premium.Group("/shares")
	{
		shares.POST("/files/:id", handlers.AdvancedShareFileController.CreateShare)
	}
}

func setupSuperAdminRoutes(superAdmin *gin.RouterGroup, handlers *SuperAdminHandlers) {
	superAdmin.GET("/me", handlers.LoginController.GetMe)
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
