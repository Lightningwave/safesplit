package main

import (
	"fmt"
	"log"
	"os"
	"safesplit/config"
	"safesplit/models"
	"safesplit/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize all required models
	userModel := models.NewUserModel(db)
	activityLogModel := models.NewActivityLogModel(db)
	fileModel := models.NewFileModel(db)

	// Initialize route handlers with all required dependencies
	handlers := routes.NewRouteHandlers(
		db,
		userModel,
		activityLogModel,
		fileModel,
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

	// Create storage directory if it doesn't exist
	if err := createStorageDirectories(); err != nil {
		log.Fatal("Failed to create storage directories:", err)
	}

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
	storagePath := "storage/encrypted"
	if err := os.MkdirAll(storagePath, 0700); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}
	return nil
}
