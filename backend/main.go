package main

import (
	"fmt"
	"log"
	"os"
	"safesplit/config"
	"safesplit/jobs"
	"safesplit/models"
	"safesplit/routes"
	"safesplit/services"
	"strconv"
	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	baseStoragePath = "storage"
	nodeCount       = 3
)
func init() {
    if err := godotenv.Load(); err != nil {  
        log.Fatal("Error loading .env file")
    }
}
func main() {
	// Initialize database connection
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// Initialize SMTP config from environment
	smtpConfig := services.SMTPConfig{
		Host: os.Getenv("SMTP_HOST"),
		Port: func() int {
			port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
			if err != nil {
				return 587
			}
			return port
		}(),
		Username:  os.Getenv("SMTP_USERNAME"),
		Password:  os.Getenv("SMTP_PASSWORD"),
		FromName:  os.Getenv("SMTP_FROM_NAME"),
		FromEmail: os.Getenv("SMTP_FROM_EMAIL"),
	}
	// Initialize email services
	emailService, err := services.NewSMTPEmailService(smtpConfig)
	if err != nil {
		log.Fatal("Failed to initialize email service:", err)
	}
	twoFactorService := services.NewTwoFactorAuthService(emailService)

	// Initialize subscription handler and scheduler
	subscriptionHandler := jobs.NewSubscriptionHandler(db)
	jobs.StartSubscriptionScheduler(subscriptionHandler)

	// Initialize distributed storage service
	storageService, err := services.NewDistributedStorageService(baseStoragePath, nodeCount)
	if err != nil {
		log.Fatal("Failed to initialize distributed storage:", err)
	}
	// Initialize subscription handler

	// Initialize server master key
	serverMasterKeyModel := models.NewServerMasterKeyModel(db)
	if err := serverMasterKeyModel.Initialize(); err != nil {
		log.Fatal("Failed to initialize server master key:", err)
	}

	// Initialize all required models
	userModel := models.NewUserModel(db, twoFactorService)
	passwordHistoryModel := models.NewPasswordHistoryModel(db)
	billingModel := models.NewBillingModel(db, userModel)
	activityLogModel := models.NewActivityLogModel(db)
	folderModel := models.NewFolderModel(db)
	fileShareModel := models.NewFileShareModel(db)
	keyFragmentModel := models.NewKeyFragmentModel(db, storageService)
	keyRotationModel := models.NewKeyRotationModel(db)

	// Initialize core services
	shamirService := services.NewShamirService(nodeCount)
	encryptionService := services.NewEncryptionService(shamirService)

	// Initialize compression service
	compressionService, err := services.NewCompressionService()
	if err != nil {
		log.Fatal("Failed to initialize compression service:", err)
	}
	defer compressionService.Close()

	// Initialize Reed-Solomon service with the same storage service
	rsService, err := services.NewReedSolomonService(storageService)
	if err != nil {
		log.Fatal("Failed to initialize Reed-Solomon service:", err)
	}

	// Initialize file model with server master key model
	fileModel := models.NewFileModel(
		db,
		rsService,
		serverMasterKeyModel,
		encryptionService,
		keyFragmentModel,
	)
	// Initialize route handlers with all required dependencies
	handlers := routes.NewRouteHandlers(
		db,
		userModel,
		billingModel,
		passwordHistoryModel,
		activityLogModel,
		fileModel,
		folderModel,
		fileShareModel,
		keyFragmentModel,
		keyRotationModel,
		serverMasterKeyModel,
		encryptionService,
		shamirService,
		compressionService,
		rsService,
		twoFactorService,
	)

	// Set up the Gin router with default middleware
	router := gin.Default()

	// Configure CORS settings for secure cross-origin requests
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
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

	// Add debug endpoint for route verification
	router.GET("/debug", func(c *gin.Context) {
		routes := []string{}
		for _, route := range router.Routes() {
			routes = append(routes, fmt.Sprintf("%s %s", route.Method, route.Path))
		}
		c.JSON(200, gin.H{
			"routes": routes,
		})
	})

	// Set up all application routes
	routes.SetupRoutes(router, handlers, userModel)

	// Log all registered routes for debugging purposes
	log.Println("=== Registered Routes ===")
	for _, route := range router.Routes() {
		log.Printf("%s %s", route.Method, route.Path)
	}
	log.Println("=======================")

	// Start the server on the specified port
	port := ":8080"
	log.Printf("Server starting on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
