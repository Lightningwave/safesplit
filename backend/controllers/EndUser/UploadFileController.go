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
	"strconv"

	"github.com/gin-gonic/gin"
)

type UploadFileController struct {
	fileModel          *models.FileModel
	userModel          *models.UserModel
	activityLogModel   *models.ActivityLogModel
	encryptionService  *services.EncryptionService
	shamirService      *services.ShamirService
	keyFragmentModel   *models.KeyFragmentModel
	compressionService *services.CompressionService
}

func NewFileController(
	fileModel *models.FileModel,
	userModel *models.UserModel,
	activityLogModel *models.ActivityLogModel,
	encryptionService *services.EncryptionService,
	shamirService *services.ShamirService,
	keyFragmentModel *models.KeyFragmentModel,
	compressionService *services.CompressionService,
) *UploadFileController {
	return &UploadFileController{
		fileModel:          fileModel,
		userModel:          userModel,
		activityLogModel:   activityLogModel,
		encryptionService:  encryptionService,
		shamirService:      shamirService,
		keyFragmentModel:   keyFragmentModel,
		compressionService: compressionService,
	}
}

func (c *UploadFileController) validateShamirParameters(n, k int) error {
	if n < k {
		return fmt.Errorf("number of shares (n) must be greater than or equal to threshold (k)")
	}
	if k < 2 {
		return fmt.Errorf("threshold (k) must be at least 2")
	}
	if n > 10 {
		return fmt.Errorf("number of shares (n) cannot exceed 10")
	}
	return nil
}

func (c *UploadFileController) Upload(ctx *gin.Context) {
	log.Printf("Starting file upload request")

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

	nShares := ctx.PostForm("shares")
	threshold := ctx.PostForm("threshold")

	n, err := strconv.Atoi(nShares)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid number of shares provided",
		})
		return
	}

	k, err := strconv.Atoi(threshold)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid threshold provided",
		})
		return
	}

	if err := c.validateShamirParameters(n, k); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Printf("Error receiving file: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "No file was provided",
		})
		return
	}

	// Check storage quota
	if !currentUser.HasAvailableStorage(fileHeader.Size) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Insufficient storage space",
		})
		return
	}

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

	// Calculate original file hash
	hash := sha256.Sum256(content)
	fileHash := base64.StdEncoding.EncodeToString(hash[:])

	// Compress content
	compressed, ratio, err := c.compressionService.Compress(content)
	if err != nil {
		log.Printf("Compression failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to compress file",
		})
		return
	}

	// Encrypt the compressed content
	encrypted, iv, salt, shares, err := c.encryptionService.EncryptFile(compressed, n, k)
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to encrypt file",
		})
		return
	}

	storageDir := filepath.Join("storage", "files")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Printf("Failed to create storage directory: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to prepare storage",
		})
		return
	}

	encryptedFileName := base64.RawURLEncoding.EncodeToString([]byte(fileHeader.Filename))
	storagePath := filepath.Join(storageDir, encryptedFileName)

	fileRecord := &models.File{
		UserID:           currentUser.ID,
		Name:             encryptedFileName,
		OriginalName:     fileHeader.Filename,
		FilePath:         storagePath,
		Size:             fileHeader.Size,
		CompressedSize:   int64(len(compressed)),
		MimeType:         fileHeader.Header.Get("Content-Type"),
		EncryptionIV:     iv,
		EncryptionSalt:   salt,
		FileHash:         fileHash,
		ShareCount:       uint(n),
		Threshold:        uint(k),
		IsCompressed:     true,
		CompressionRatio: ratio,
	}

	if err := os.WriteFile(storagePath, encrypted, 0600); err != nil {
		log.Printf("Failed to save encrypted file: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to save file",
		})
		return
	}

	if err := c.fileModel.CreateFileWithFragments(fileRecord, shares, c.keyFragmentModel); err != nil {
		os.Remove(storagePath)
		log.Printf("Database error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Failed to save file information: %v", err),
		})
		return
	}

	// Update user's storage usage using UserModel
	if err := c.userModel.UpdateUserStorage(currentUser.ID, fileRecord.Size); err != nil {
		os.Remove(storagePath)
		log.Printf("Failed to update storage usage: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to update storage usage",
		})
		return
	}

	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       currentUser.ID,
		ActivityType: "upload",
		FileID:       &fileRecord.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      fmt.Sprintf("File compressed to %.2f%% of original size", ratio*100),
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	log.Printf("File upload successful: %s with %d shares and threshold of %d", fileRecord.Name, n, k)
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File uploaded successfully",
		"data": gin.H{
			"file": fileRecord,
			"compressionStats": gin.H{
				"originalSize":     fileHeader.Size,
				"compressedSize":   len(compressed),
				"compressionRatio": fmt.Sprintf("%.2f%%", ratio*100),
			},
		},
	})
}
