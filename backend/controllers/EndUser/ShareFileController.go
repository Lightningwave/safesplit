package EndUser

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
)

type ShareFileController struct {
	fileModel         *models.FileModel
	fileShareModel    *models.FileShareModel
	keyFragmentModel  *models.KeyFragmentModel
	encryptionService *services.EncryptionService
	activityLogModel  *models.ActivityLogModel
}

func NewShareFileController(
	fileModel *models.FileModel,
	fileShareModel *models.FileShareModel,
	keyFragmentModel *models.KeyFragmentModel,
	encryptionService *services.EncryptionService,
	activityLogModel *models.ActivityLogModel,
) *ShareFileController {
	return &ShareFileController{
		fileModel:         fileModel,
		fileShareModel:    fileShareModel,
		keyFragmentModel:  keyFragmentModel,
		encryptionService: encryptionService,
		activityLogModel:  activityLogModel,
	}
}

type CreateShareRequest struct {
	FileID   uint   `json:"file_id" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type AccessShareRequest struct {
	Password string `json:"password" binding:"required"`
}

// CreateShare generates a share link for a file
func (c *ShareFileController) CreateShare(ctx *gin.Context) {
	var req CreateShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	// Get current user
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	// Verify file ownership
	file, err := c.fileModel.GetFileForDownload(req.FileID, currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found or access denied",
		})
		return
	}

	// Get key fragments
	fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
	if err != nil || len(fragments) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve key fragments",
		})
		return
	}

	// Encrypt fragment with share password
	encryptedFragment, err := c.encryptionService.EncryptKeyFragment(
		fragments[0].EncryptedFragment,
		[]byte(req.Password),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process share encryption",
		})
		return
	}

	// Create share record
	share := &models.FileShare{
		FileID:               file.ID,
		SharedBy:             currentUser.ID,
		EncryptedKeyFragment: encryptedFragment,
		IsActive:             true,
	}

	if err := c.fileShareModel.CreateFileShare(share, req.Password); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to create share",
		})
		return
	}

	// Log share creation
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       currentUser.ID,
		ActivityType: "share",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
	}); err != nil {
		log.Printf("Failed to log share activity: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"share_link": share.ShareLink,
		},
	})
}

// AccessShare handles accessing a shared file
func (c *ShareFileController) AccessShare(ctx *gin.Context) {
	shareLink := ctx.Param("shareLink")
	var req AccessShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid password",
		})
		return
	}

	// Validate share access
	share, err := c.fileShareModel.ValidateShare(shareLink, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid share link or password",
		})
		return
	}

	// Get file fragments
	fragments, err := c.keyFragmentModel.GetKeyFragments(share.FileID)
	if err != nil || len(fragments) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process file access",
		})
		return
	}

	// Decrypt shared fragment
	decryptedFragment, err := c.encryptionService.DecryptKeyFragment(
		share.EncryptedKeyFragment,
		[]byte(req.Password),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process file decryption",
		})
		return
	}

	// Get file details
	file, err := c.fileModel.GetFileByID(share.FileID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found",
		})
		return
	}

	// Prepare key shares
	shares := make([]services.KeyShare, len(fragments))
	for i, fragment := range fragments {
		if i == 0 {
			shares[i] = services.KeyShare{
				Index: fragment.FragmentIndex,
				Value: string(decryptedFragment),
			}
		} else {
			shares[i] = services.KeyShare{
				Index: fragment.FragmentIndex,
				Value: string(fragment.EncryptedFragment),
			}
		}
	}

	// Read and decrypt file
	encryptedData, err := c.fileModel.ReadFileContent(file.FilePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to read file",
		})
		return
	}

	decryptedData, err := c.encryptionService.DecryptFile(
		encryptedData,
		file.EncryptionIV,
		shares,
		int(file.Threshold),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to decrypt file",
		})
		return
	}

	// Log share access (only once)
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       share.SharedBy,
		ActivityType: "download",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      "Shared file download",
	}); err != nil {
		log.Printf("Failed to log share download activity: %v", err)
	}

	// Improved filename handling for download
	sanitizedFilename := strings.ReplaceAll(file.OriginalName, `"`, `\"`)
	encodedFilename := url.QueryEscape(sanitizedFilename)

	// Set headers with both encoded and unencoded filenames for better browser compatibility
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Type, Content-Length")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
		sanitizedFilename,
		encodedFilename))
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(decryptedData)))

	// Send file
	ctx.Data(http.StatusOK, file.MimeType, decryptedData)
}
