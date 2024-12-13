package EndUser

import (
	"fmt"
	"log"

	//"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"safesplit/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UploadFileController struct {
	db               *gorm.DB
	fileModel        *models.FileModel
	activityLogModel *models.ActivityLogModel
}

func NewFileController(
	db *gorm.DB,
	fileModel *models.FileModel,
	activityLogModel *models.ActivityLogModel,
) *UploadFileController {
	return &UploadFileController{
		db:               db,
		fileModel:        fileModel,
		activityLogModel: activityLogModel,
	}
}

func (c *UploadFileController) Upload(ctx *gin.Context) {
	log.Printf("Starting file upload request")

	// Retrieve the authenticated user from context
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Invalid user data",
		})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		log.Printf("Error receiving file: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Error receiving file: %v", err),
		})
		return
	}

	log.Printf("Received file: %s, size: %d bytes", file.Filename, file.Size)

	// Ensure storage directory exists
	storageDir := filepath.Join("storage", "files")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Printf("Failed to create storage directory: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Storage error: %v", err),
		})
		return
	}

	// Generate storage path
	storagePath := filepath.Join(storageDir, file.Filename)
	log.Printf("Storage path: %s", storagePath)

	// Begin transaction
	tx := c.db.Begin()
	log.Printf("Started database transaction")

	// Create file record with authenticated user ID
	fileRecord := &models.File{
		UserID:       currentUser.ID, // Use authenticated user ID
		Name:         file.Filename,
		OriginalName: file.Filename,
		FilePath:     storagePath,
		Size:         file.Size,
		MimeType:     file.Header.Get("Content-Type"),
	}

	if err := tx.Create(fileRecord).Error; err != nil {
		tx.Rollback()
		log.Printf("Database error creating file record: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Database error: %v", err),
		})
		return
	}

	// Save file
	if err := ctx.SaveUploadedFile(file, storagePath); err != nil {
		tx.Rollback()
		log.Printf("Error saving file: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("File save error: %v", err),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		os.Remove(storagePath) // Clean up file if transaction fails
		log.Printf("Transaction commit error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Transaction error: %v", err),
		})
		return
	}

	log.Printf("File upload successful: %s", fileRecord.Name)
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File uploaded successfully",
		"data": gin.H{
			"file": fileRecord,
		},
	})
}
