package EndUser

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
)

type UploadFileController struct {
	fileModel         *models.FileModel
	activityLogModel  *models.ActivityLogModel
	encryptionService *services.EncryptionService
	shamirService     *services.ShamirService
	keyFragmentModel  *models.KeyFragmentModel
}

func NewFileController(
	fileModel *models.FileModel,
	activityLogModel *models.ActivityLogModel,
	encryptionService *services.EncryptionService,
	shamirService *services.ShamirService,
	keyFragmentModel *models.KeyFragmentModel,
) *UploadFileController {
	return &UploadFileController{
		fileModel:         fileModel,
		activityLogModel:  activityLogModel,
		encryptionService: encryptionService,
		shamirService:     shamirService,
		keyFragmentModel:  keyFragmentModel,
	}
}

func (c *UploadFileController) Upload(ctx *gin.Context) {
	log.Printf("Starting file upload request")

	// Get authenticated user
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

	// Get uploaded file
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Printf("Error receiving file: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "No file was provided",
		})
		return
	}

	// Read file content
	src, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to read file",
		})
		return
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to read file content",
		})
		return
	}

	// Calculate file hash
	hash := sha256.Sum256(content)
	fileHash := base64.StdEncoding.EncodeToString(hash[:])

	// Define n and k for Shamir's Secret Sharing
	n := 2
	k := 2

	// Encrypt file
	encrypted, iv, salt, shares, err := c.encryptionService.EncryptFile(content, n, k)
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to encrypt file",
		})
		return
	}

	// Ensure storage directory exists
	storageDir := filepath.Join("storage", "files")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Printf("Failed to create storage directory: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to prepare storage",
		})
		return
	}

	// Generate unique filename
	encryptedFileName := base64.RawURLEncoding.EncodeToString([]byte(fileHeader.Filename))
	storagePath := filepath.Join(storageDir, encryptedFileName)

	// Create file record
	fileRecord := &models.File{
		UserID:         currentUser.ID,
		Name:           encryptedFileName,
		OriginalName:   fileHeader.Filename,
		FilePath:       storagePath,
		Size:           fileHeader.Size,
		MimeType:       fileHeader.Header.Get("Content-Type"),
		EncryptionIV:   iv,
		EncryptionSalt: salt,
		FileHash:       fileHash,
	}

	// Save encrypted file
	if err := os.WriteFile(storagePath, encrypted, 0600); err != nil {
		log.Printf("Failed to save encrypted file: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to save file",
		})
		return
	}

	// Create database records with key fragments
	if err := c.fileModel.CreateFileWithFragments(fileRecord, shares, c.keyFragmentModel); err != nil {
		os.Remove(storagePath)
		log.Printf("Database error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Failed to save file information: %v", err),
		})
		return
	}

	// Log activity
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       currentUser.ID,
		ActivityType: "upload",
		FileID:       &fileRecord.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
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
