package EndUser

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
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
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "Unauthorized access"})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Invalid user data"})
		return
	}

	// Get current folder ID (optional)
	var folderID *uint
	if folderIDStr := ctx.PostForm("folder_id"); folderIDStr != "" {
		id, err := strconv.ParseUint(folderIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid folder ID format"})
			return
		}
		parsedID := uint(id)
		folderID = &parsedID
	}

	nShares := ctx.PostForm("shares")
	threshold := ctx.PostForm("threshold")

	n, err := strconv.Atoi(nShares)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid number of shares provided"})
		return
	}

	k, err := strconv.Atoi(threshold)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid threshold provided"})
		return
	}

	if err := c.validateShamirParameters(n, k); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Printf("Error receiving file: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "No file was provided"})
		return
	}

	// Check storage quota
	if !currentUser.HasAvailableStorage(fileHeader.Size) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Insufficient storage space"})
		return
	}

	// Process file upload
	processedFile, err := c.processFileUpload(fileHeader, n, k)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Create file record
	encryptedFileName := base64.RawURLEncoding.EncodeToString([]byte(fileHeader.Filename))
	fileRecord := &models.File{
		UserID:           currentUser.ID,
		FolderID:         folderID, // Add this line to set the folder ID
		Name:             encryptedFileName,
		OriginalName:     fileHeader.Filename,
		Size:             fileHeader.Size,
		CompressedSize:   int64(len(processedFile.compressed)),
		MimeType:         fileHeader.Header.Get("Content-Type"),
		EncryptionIV:     processedFile.iv,
		EncryptionSalt:   processedFile.salt,
		FileHash:         processedFile.fileHash,
		ShareCount:       uint(n),
		Threshold:        uint(k),
		IsCompressed:     true,
		CompressionRatio: processedFile.ratio,
	}

	// Get storage path from model
	storagePath := c.fileModel.GenerateStoragePath(encryptedFileName)

	// Save encrypted file
	if err := os.WriteFile(storagePath, processedFile.encrypted, 0600); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Failed to save file"})
		return
	}

	// Create file record with fragments
	if err := c.fileModel.CreateFileWithFragments(fileRecord, processedFile.shares, c.keyFragmentModel); err != nil {
		os.Remove(storagePath)
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
		Details:      fmt.Sprintf("File uploaded and compressed to %.2f%% of original size", processedFile.ratio*100),
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File uploaded successfully",
		"data": gin.H{
			"file": fileRecord,
			"compressionStats": gin.H{
				"originalSize":     fileRecord.Size,
				"compressedSize":   fileRecord.CompressedSize,
				"compressionRatio": fmt.Sprintf("%.2f%%", processedFile.ratio*100),
			},
			"folder_id": folderID,
		},
	})
}

type processedFile struct {
	compressed []byte
	encrypted  []byte
	iv         []byte
	salt       []byte
	shares     []services.KeyShare
	fileHash   string
	ratio      float64
}

func (c *UploadFileController) processFileUpload(fileHeader *multipart.FileHeader, n, k int) (*processedFile, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Calculate hash
	hash := sha256.Sum256(content)
	fileHash := base64.StdEncoding.EncodeToString(hash[:])

	// Compress content
	compressed, ratio, err := c.compressionService.Compress(content)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Encrypt the compressed content
	encrypted, iv, salt, shares, err := c.encryptionService.EncryptFile(compressed, n, k)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return &processedFile{
		compressed: compressed,
		encrypted:  encrypted,
		iv:         iv,
		salt:       salt,
		shares:     shares,
		fileHash:   fileHash,
		ratio:      ratio,
	}, nil
}
