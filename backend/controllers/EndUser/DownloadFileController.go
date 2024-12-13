package EndUser

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"safesplit/models"
	"safesplit/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DownloadFileController struct {
	fileModel         *models.FileModel
	keyFragmentModel  *models.KeyFragmentModel
	encryptionService *services.EncryptionService
	activityLogModel  *models.ActivityLogModel
}

func NewDownloadFileController(
	fileModel *models.FileModel,
	keyFragmentModel *models.KeyFragmentModel,
	encryptionService *services.EncryptionService,
	activityLogModel *models.ActivityLogModel,
) *DownloadFileController {
	return &DownloadFileController{
		fileModel:         fileModel,
		keyFragmentModel:  keyFragmentModel,
		encryptionService: encryptionService,
		activityLogModel:  activityLogModel,
	}
}

func (c *DownloadFileController) Download(ctx *gin.Context) {
	log.Printf("Starting file download request")

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

	fileID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		log.Printf("Invalid file ID: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid file ID",
		})
		return
	}

	// Fetch the file record
	file, err := c.fileModel.GetFileForDownload(uint(fileID), currentUser.ID)
	if err != nil {
		log.Printf("Error fetching file: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "file not found or access denied" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// Check if file exists
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		log.Printf("File not found on server: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found on server",
		})
		return
	}

	// Read encrypted file
	encryptedData, err := os.ReadFile(file.FilePath)
	if err != nil {
		log.Printf("Failed to read encrypted file: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to read encrypted file",
		})
		return
	}

	// Fetch key fragments
	fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
	if err != nil {
		log.Printf("Failed to retrieve key fragments: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve key fragments",
		})
		return
	}

	// Convert fragments to KeyShares
	shares := make([]services.KeyShare, len(fragments))
	for i, fragment := range fragments {
		shares[i] = services.KeyShare{
			Index: fragment.FragmentIndex,
			Value: fragment.EncryptedFragment,
		}
		log.Printf("Processing share %d - Index: %d, Value length: %d",
			i, fragment.FragmentIndex, len(fragment.EncryptedFragment))
	}

	// Decrypt file - using k=2 to match upload configuration
	decryptedData, err := c.encryptionService.DecryptFile(encryptedData, file.EncryptionIV, shares, 2)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Failed to decrypt file: %v", err),
		})
		return
	}

	// Log activity
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       currentUser.ID,
		ActivityType: "download",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	// Set response headers
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(decryptedData)))

	log.Printf("File download successful: %s", file.Name)
	ctx.Data(http.StatusOK, file.MimeType, decryptedData)
}
