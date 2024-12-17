package EndUser

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	FileID       uint       `json:"file_id" binding:"required"`
	Password     string     `json:"password" binding:"required,min=6"`
	ExpiresAt    *time.Time `json:"expires_at"`
	MaxDownloads *int       `json:"max_downloads"`
}

type AccessShareRequest struct {
	Password string `json:"password" binding:"required"`
}

func (c *ShareFileController) CreateShare(ctx *gin.Context) {
	var req CreateShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	file, err := c.fileModel.GetFileForDownload(req.FileID, currentUser.ID)
	if err != nil {
		log.Printf("File access error: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found or access denied",
		})
		return
	}

	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Expiry date cannot be in the past",
		})
		return
	}

	fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
	if err != nil || len(fragments) == 0 {
		log.Printf("Failed to retrieve key fragments: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve key fragments",
		})
		return
	}

	encryptedFragment, err := c.encryptionService.EncryptKeyFragment(
		fragments[0].EncryptedFragment,
		[]byte(req.Password),
	)
	if err != nil {
		log.Printf("Fragment encryption error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process share encryption",
		})
		return
	}

	share := &models.FileShare{
		FileID:               file.ID,
		SharedBy:             currentUser.ID,
		EncryptedKeyFragment: encryptedFragment,
		ExpiresAt:            req.ExpiresAt,
		MaxDownloads:         req.MaxDownloads,
		IsActive:             true,
	}

	if err := c.fileShareModel.CreateFileShareWithStatus(share, req.Password); err != nil {
		log.Printf("Share creation error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to create share",
		})
		return
	}

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

	share, err := c.fileShareModel.ValidateShareAccess(shareLink, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	allFragments, err := c.keyFragmentModel.GetKeyFragments(share.FileID)
	if err != nil || len(allFragments) == 0 {
		log.Printf("Failed to retrieve file fragments: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process file access",
		})
		return
	}

	decryptedFragment, err := c.encryptionService.DecryptKeyFragment(
		share.EncryptedKeyFragment,
		[]byte(req.Password),
	)
	if err != nil {
		log.Printf("Fragment decryption error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process file decryption",
		})
		return
	}

	file, err := c.fileModel.GetFileByID(share.FileID)
	if err != nil {
		log.Printf("File retrieval error: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found",
		})
		return
	}

	shares := make([]services.KeyShare, len(allFragments))
	for i, fragment := range allFragments {
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

	encryptedData, err := c.fileModel.ReadFileContent(file.FilePath)
	if err != nil {
		log.Printf("File read error: %v", err)
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
		log.Printf("File decryption error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to decrypt file",
		})
		return
	}

	if err := c.fileShareModel.IncrementDownloadCount(share.ID); err != nil {
		log.Printf("Failed to increment download count: %v", err)
	}

	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		ActivityType: "share_download",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
	}); err != nil {
		log.Printf("Failed to log share download activity: %v", err)
	}

	// Sanitize filename and prepare headers
	sanitizedFilename := strings.ReplaceAll(file.OriginalName, `"`, `\"`)

	// Set response headers
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Type, Content-Length")
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, sanitizedFilename))
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(decryptedData)))
	ctx.Header("X-Original-Filename", url.QueryEscape(file.OriginalName))

	// Send the decrypted file data as a downloadable attachment
	ctx.Data(http.StatusOK, file.MimeType, decryptedData)

}
