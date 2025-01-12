package main

import (
	"fmt"
	"log"
	"os"
	"safesplit/config"
	"safesplit/models"
	"safesplit/routes"
	"safesplit/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create storage directories
	if err := createStorageDirectories(); err != nil {
		log.Fatal("Failed to create storage directories:", err)
	}

	// Initialize all required models
	userModel := models.NewUserModel(db)
	passwordHistoryModel := models.NewPasswordHistoryModel(db)
	billingModel := models.NewBillingModel(db, userModel)
	activityLogModel := models.NewActivityLogModel(db)
	folderModel := models.NewFolderModel(db)
	fileShareModel := models.NewFileShareModel(db)
	keyFragmentModel := models.NewKeyFragmentModel(db)

	// Initialize core services
	shamirService := services.NewShamirService()
	encryptionService := services.NewEncryptionService(shamirService)

	// Initialize compression service
	compressionService, err := services.NewCompressionService()
	if err != nil {
		log.Fatal("Failed to initialize compression service:", err)
	}
	defer compressionService.Close()

	// Initialize Reed-Solomon service with distributed storage
	rsService, err := services.NewReedSolomonService("storage/shards", 3) // Using 3 storage nodes
	if err != nil {
		log.Fatal("Failed to initialize Reed-Solomon service:", err)
	}
	fileModel := models.NewFileModel(db, rsService)

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
		encryptionService,
		shamirService,
		compressionService,
		rsService,
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

// createStorageDirectories ensures the required storage directories exist
func createStorageDirectories() error {
	// Create main storage directories
	paths := []string{
		"storage/encrypted",
		"storage/files",
		"storage/shards",
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0700); err != nil {
			return fmt.Errorf("failed to create storage directory %s: %w", path, err)
		}
	}

	return nil
}
