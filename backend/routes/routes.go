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
	TwoFactorController       *EndUser.TwoFactorController
	SuperAdminHandlers        *SuperAdminHandlers
	SysAdminHandlers          *SysAdminHandlers
	EndUserHandlers           *EndUserHandlers
	PremiumUserHandlers       *PremiumUserHandlers
}

type EndUserHandlers struct {
	UploadFileController     *EndUser.UploadFileController
	MassUploadController     *EndUser.MassUploadFileController
	ViewFilesController      *EndUser.ViewFilesController
	DownloadFileController   *EndUser.DownloadFileController
	MassDownloadController   *EndUser.MassDownloadFileController
	DeleteFileController     *EndUser.DeleteFileController
	MassDeleteFileController *EndUser.MassDeleteFileController
	ArchiveFileController    *EndUser.ArchiveFileController
	UnarchiveFileController  *EndUser.UnarchiveFileController
	MassArchiveController    *EndUser.MassArchiveFileController
	MassUnarchiveController  *EndUser.MassUnarchiveFileController
	ShareFileController      *EndUser.ShareFileController
	CreateFolderController   *EndUser.CreateFolderController
	ViewFolderController     *EndUser.ViewFolderController
	DeleteFolderController   *EndUser.DeleteFolderController
	PasswordResetController  *EndUser.PasswordResetController
	ViewStorageController    *EndUser.ViewStorageController
	PaymentController        *EndUser.PaymentController
	SubscriptionController   *EndUser.SubscriptionController
	ReportController         *EndUser.ReportController
	FeedbackController       *EndUser.FeedbackController
}
type PremiumUserHandlers struct {
	FileRecoveryController      *PremiumUser.FileRecoveryController
	AdvancedShareFileController *PremiumUser.ShareFileController
	UpdateBillingController     *PremiumUser.UpdateBillingController
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
	ViewFeedbacksController          *SysAdmin.ViewFeedbacksController
	ViewReportsController            *SysAdmin.ViewReportsController
	ViewBillingRecordsController     *SysAdmin.ViewBillingRecordsController
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
	serverMasterKeyModel *models.ServerMasterKeyModel,
	feedbackModel *models.FeedbackModel,
	encryptionService *services.EncryptionService,
	shamirService *services.ShamirService,
	compressionService *services.CompressionService,
	rsService *services.ReedSolomonService,
	twoFactorService *services.TwoFactorAuthService,
	emailService *services.SMTPEmailService,
) *RouteHandlers {
	superAdminLoginController := SuperAdmin.NewLoginController(userModel)
	return &RouteHandlers{
		LoginController:           controllers.NewLoginController(userModel, billingModel),
		SuperAdminLoginController: superAdminLoginController,
		CreateAccountController:   controllers.NewCreateAccountController(userModel, passwordHistoryModel),
		TwoFactorController:       EndUser.NewTwoFactorController(userModel, twoFactorService),
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
			ViewFeedbacksController:          SysAdmin.NewViewFeedbacksController(feedbackModel),
			ViewReportsController:            SysAdmin.NewViewReportsController(feedbackModel, userModel),
			ViewBillingRecordsController:     SysAdmin.NewViewBillingRecordsController(billingModel),
		},
		EndUserHandlers: &EndUserHandlers{
			UploadFileController:     EndUser.NewFileController(fileModel, userModel, activityLogModel, encryptionService, shamirService, keyFragmentModel, compressionService, folderModel, rsService, serverMasterKeyModel),
			MassUploadController:     EndUser.NewMassUploadFileController(fileModel, userModel, activityLogModel, encryptionService, shamirService, keyFragmentModel, compressionService, folderModel, rsService, serverMasterKeyModel),
			ViewFilesController:      EndUser.NewViewFilesController(fileModel, folderModel),
			DownloadFileController:   EndUser.NewDownloadFileController(fileModel, keyFragmentModel, encryptionService, activityLogModel, compressionService, rsService, serverMasterKeyModel),
			MassDownloadController:   EndUser.NewMassDownloadFileController(fileModel, keyFragmentModel, encryptionService, activityLogModel, compressionService, rsService, serverMasterKeyModel),
			DeleteFileController:     EndUser.NewDeleteFileController(fileModel),
			MassDeleteFileController: EndUser.NewMassDeleteFileController(fileModel),
			ArchiveFileController:    EndUser.NewArchiveFileController(fileModel),
			UnarchiveFileController:  EndUser.NewUnarchiveFileController(fileModel),
			MassArchiveController:    EndUser.NewMassArchiveFileController(fileModel),
			MassUnarchiveController:  EndUser.NewMassUnarchiveFileController(fileModel),
			ShareFileController:      EndUser.NewShareFileController(fileModel, fileShareModel, keyFragmentModel, encryptionService, activityLogModel, rsService, userModel, serverMasterKeyModel, twoFactorService, emailService, compressionService),
			CreateFolderController:   EndUser.NewCreateFolderController(folderModel, activityLogModel),
			ViewFolderController:     EndUser.NewViewFolderController(folderModel, fileModel),
			DeleteFolderController:   EndUser.NewDeleteFolderController(folderModel, activityLogModel),
			PasswordResetController:  EndUser.NewPasswordResetController(userModel, passwordHistoryModel, keyFragmentModel, fileModel),
			ViewStorageController:    EndUser.NewViewStorageController(fileModel, userModel),
			PaymentController:        EndUser.NewPaymentController(billingModel),
			SubscriptionController:   EndUser.NewSubscriptionController(billingModel),
			ReportController:         EndUser.NewReportController(feedbackModel, fileModel),
			FeedbackController:       EndUser.NewFeedbackController(feedbackModel),
		},
		PremiumUserHandlers: &PremiumUserHandlers{
			FileRecoveryController:      PremiumUser.NewFileRecoveryController(fileModel),
			AdvancedShareFileController: PremiumUser.NewShareFileController(fileModel, fileShareModel, keyFragmentModel, encryptionService, activityLogModel, rsService, userModel, serverMasterKeyModel, twoFactorService, emailService, compressionService),
			UpdateBillingController:     PremiumUser.NewUpdateBillingController(billingModel),
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

	// Public share routes
	api.GET("/files/share/:shareLink", handlers.EndUserHandlers.ShareFileController.AccessShare)
	api.POST("/files/share/:shareLink", handlers.EndUserHandlers.ShareFileController.AccessShare)
	api.POST("/files/share/:shareLink/verify", handlers.EndUserHandlers.ShareFileController.Verify2FAAndDownload)

	// Premium share routes
	api.GET("/premium/shares/:shareLink", handlers.PremiumUserHandlers.AdvancedShareFileController.AccessShare)
	api.POST("/premium/shares/:shareLink", handlers.PremiumUserHandlers.AdvancedShareFileController.AccessShare)
	api.POST("/premium/shares/:shareLink/verify", handlers.PremiumUserHandlers.AdvancedShareFileController.Verify2FAAndDownload)

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func setupProtectedRoutes(protected *gin.RouterGroup, handlers *RouteHandlers) {
	protected.GET("/me", handlers.LoginController.GetMe)

	// 2FA routes
	twoFactor := protected.Group("/2fa")
	{
		twoFactor.GET("/status", handlers.TwoFactorController.GetTwoFactorStatus)
		twoFactor.POST("/enable/initiate", handlers.TwoFactorController.InitiateEnable2FA)
		twoFactor.POST("/enable/verify", handlers.TwoFactorController.VerifyAndEnable2FA)
		twoFactor.POST("/disable/initiate", handlers.TwoFactorController.InitiateDisable2FA)
		twoFactor.POST("/disable/verify", handlers.TwoFactorController.VerifyAndDisable2FA)
	}

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
	files := protected.Group("/files")
	{
		files.GET("", handlers.ViewFilesController.ListUserFiles)
		files.GET("/:id/download", handlers.DownloadFileController.Download)
		files.POST("/mass-download", handlers.MassDownloadController.MassDownload)
		files.GET("/mass-download/:id", handlers.MassDownloadController.GetFile)
		files.POST("/upload", handlers.UploadFileController.Upload)
		files.POST("/mass-upload", handlers.MassUploadController.MassUpload)
		files.GET("/encryption/options", handlers.UploadFileController.GetEncryptionOptions)
		files.DELETE("/:id", handlers.DeleteFileController.Delete)
		files.POST("/mass-delete", handlers.MassDeleteFileController.Delete)
		files.PUT("/:id/archive", handlers.ArchiveFileController.Archive)
		files.PUT("/:id/unarchive", handlers.UnarchiveFileController.Unarchive)
		files.POST("/mass-archive", handlers.MassArchiveController.Archive)
		files.POST("/mass-unarchive", handlers.MassUnarchiveController.Unarchive)
		files.POST("/:id/share", handlers.ShareFileController.CreateShare)
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
	payment := protected.Group("/payment")
	{
		payment.POST("/upgrade", handlers.PaymentController.ProcessPayment)
		payment.GET("/status", handlers.PaymentController.GetPaymentStatus)
		payment.POST("/cancel", handlers.SubscriptionController.CancelSubscription)
	}
	feedback := protected.Group("/feedback")
	{
		feedback.POST("", handlers.FeedbackController.SubmitFeedback)
		feedback.GET("", handlers.FeedbackController.GetUserFeedback)
		feedback.GET("/categories", handlers.FeedbackController.GetFeedbackCategories)
	}

	reports := protected.Group("/reports")
	{
		reports.POST("/file/:id", handlers.ReportController.ReportFile)
		reports.POST("/share/:shareLink", handlers.ReportController.ReportShare)
		reports.GET("", handlers.ReportController.GetUserReports)
	}

}
func setupPremiumUserRoutes(premium *gin.RouterGroup, handlers *PremiumUserHandlers) {

	recovery := premium.Group("/recovery")
	{
		recovery.GET("/files", handlers.FileRecoveryController.ListRecoverableFiles)
		recovery.POST("/files/:fileId", handlers.FileRecoveryController.RecoverFile)
	}
	shares := premium.Group("/shares")
	{
		shares.POST("/files/:id", handlers.AdvancedShareFileController.CreateShare)
	}
	billing := premium.Group("/billing")
	{
		billing.GET("/details", handlers.UpdateBillingController.GetBillingDetails)
		billing.PUT("/details", handlers.UpdateBillingController.UpdateBillingDetails)
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

	feedback := sysAdmin.Group("/feedback")
	{
		feedback.GET("", handlers.ViewFeedbacksController.GetAllFeedbacks)
		feedback.GET("/:id", handlers.ViewFeedbacksController.GetFeedback)
		feedback.PUT("/:id/status", handlers.ViewFeedbacksController.UpdateFeedbackStatus)
		feedback.GET("/stats", handlers.ViewFeedbacksController.GetFeedbackStats)
	}

	reports := sysAdmin.Group("/reports")
	{
		reports.GET("", handlers.ViewReportsController.GetAllReports)
		reports.GET("/:id", handlers.ViewReportsController.GetReportDetails)
		reports.PUT("/:id/status", handlers.ViewReportsController.UpdateReportStatus)
		reports.GET("/stats", handlers.ViewReportsController.GetReportStats)
	}
	billing := sysAdmin.Group("/billing")
	{
		billing.GET("/records", handlers.ViewBillingRecordsController.GetAllBillingRecords)
		billing.GET("/stats", handlers.ViewBillingRecordsController.GetBillingStats)
		billing.GET("/expiring", handlers.ViewBillingRecordsController.GetExpiringSubscriptions)
	}
}
