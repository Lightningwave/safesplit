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
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	nodeCount = 3
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

	// Initialize job handler and scheduler
	jobManager := jobs.NewJobManager(db)
	jobManager.StartAllJobs()
	// Initialize S3 storage service
	s3Configs := []struct {
		Region     string
		BucketName string
	}{
		{
			Region:     os.Getenv("S3_REGION_1"),
			BucketName: os.Getenv("S3_BUCKET_1"),
		},
		{
			Region:     os.Getenv("S3_REGION_2"),
			BucketName: os.Getenv("S3_BUCKET_2"),
		},
		{
			Region:     os.Getenv("S3_REGION_3"),
			BucketName: os.Getenv("S3_BUCKET_3"),
		},
	}

	// Validate S3 configuration
	for i, cfg := range s3Configs {
		if cfg.Region == "" {
			log.Fatalf("S3_REGION_%d is required", i+1)
		}
		if cfg.BucketName == "" {
			log.Fatalf("S3_BUCKET_%d is required", i+1)
		}
	}

	storageService, err := services.NewMultiS3StorageService(s3Configs)
	if err != nil {
		log.Fatal("Failed to initialize S3 storage:", err)
	}

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
	feedbackModel := models.NewFeedbackModel(db)

	// Initialize core services
	shamirService := services.NewShamirService(nodeCount)
	encryptionService := services.NewEncryptionService(shamirService)

	// Initialize compression service
	compressionService, err := services.NewCompressionService()
	if err != nil {
		log.Fatal("Failed to initialize compression service:", err)
	}
	defer compressionService.Close()

	// Initialize Reed-Solomon service with S3 storage
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

	// Start cleanup scheduler for deleted files
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		log.Println("Starting file cleanup scheduler...")
		for {
			select {
			case <-ticker.C:
				log.Println("Starting scheduled cleanup of old deleted files...")
				if err := fileModel.CleanupOldDeletedFiles(30); err != nil {
					log.Printf("Error during scheduled file cleanup: %v", err)
				} else {
					log.Println("Completed scheduled cleanup of old deleted files")
				}
			}
		}
	}()

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
		serverMasterKeyModel,
		feedbackModel,
		encryptionService,
		shamirService,
		compressionService,
		rsService,
		twoFactorService,
		emailService,
	)

	// Set up the Gin router with default middleware
	router := gin.Default()

	// Configure CORS settings for secure cross-origin requests
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"https://safesplit.xyz/", "http://35.240.140.29/"}
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
