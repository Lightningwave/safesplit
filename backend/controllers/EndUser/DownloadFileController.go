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
	fileModel          *models.FileModel
	keyFragmentModel   *models.KeyFragmentModel
	encryptionService  *services.EncryptionService
	activityLogModel   *models.ActivityLogModel
	compressionService *services.CompressionService
}

func NewDownloadFileController(
	fileModel *models.FileModel,
	keyFragmentModel *models.KeyFragmentModel,
	encryptionService *services.EncryptionService,
	activityLogModel *models.ActivityLogModel,
	compressionService *services.CompressionService,
) *DownloadFileController {
	return &DownloadFileController{
		fileModel:          fileModel,
		keyFragmentModel:   keyFragmentModel,
		encryptionService:  encryptionService,
		activityLogModel:   activityLogModel,
		compressionService: compressionService,
	}
}

func (c *DownloadFileController) Download(ctx *gin.Context) {
	log.Printf("Starting file download request")

	// Get authenticated user
	currentUser, err := c.getCurrentUser(ctx)
	if err != nil {
		return // Error response is handled in getCurrentUser
	}

	// Parse and validate file ID
	fileID, err := c.getFileID(ctx)
	if err != nil {
		return // Error response is handled in getFileID
	}

	// Fetch and validate file
	file, err := c.getAndValidateFile(ctx, currentUser.ID, fileID)
	if err != nil {
		return // Error response is handled in getAndValidateFile
	}

	// Process file data
	finalData, err := c.processFileData(ctx, file)
	if err != nil {
		return // Error response is handled in processFileData
	}

	// Log activity
	c.logDownloadActivity(currentUser, file, ctx.ClientIP())

	// Send file to client
	c.sendFileResponse(ctx, file, finalData)
}

func (c *DownloadFileController) getCurrentUser(ctx *gin.Context) (*models.User, error) {
	user, exists := ctx.Get("user")
	if !exists {
		log.Printf("User authentication failed - user not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return nil, fmt.Errorf("unauthorized")
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		log.Printf("User authentication failed - invalid user type in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Invalid user data",
		})
		return nil, fmt.Errorf("invalid user data")
	}

	return currentUser, nil
}

func (c *DownloadFileController) getFileID(ctx *gin.Context) (uint, error) {
	fileID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		log.Printf("Invalid file ID: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid file ID",
		})
		return 0, err
	}
	return uint(fileID), nil
}

func (c *DownloadFileController) getAndValidateFile(ctx *gin.Context, userID, fileID uint) (*models.File, error) {
	file, err := c.fileModel.GetFileForDownload(fileID, userID)
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
		return nil, err
	}

	if err := c.validateFile(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}

func (c *DownloadFileController) validateFile(ctx *gin.Context, file *models.File) error {
	if file.Threshold < 2 {
		log.Printf("Invalid threshold value in file record: %d", file.Threshold)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Invalid file configuration",
		})
		return fmt.Errorf("invalid threshold")
	}

	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		log.Printf("Physical file not found at path: %s", file.FilePath)
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found on server",
		})
		return err
	}

	return nil
}

func (c *DownloadFileController) processFileData(ctx *gin.Context, file *models.File) ([]byte, error) {
	// Read encrypted file
	encryptedData, err := os.ReadFile(file.FilePath)
	if err != nil {
		log.Printf("Failed to read encrypted file: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to read encrypted file",
		})
		return nil, err
	}

	// Get and validate key fragments
	shares, err := c.getKeyShares(ctx, file)
	if err != nil {
		return nil, err
	}

	// Decrypt file
	decryptedData, err := c.encryptionService.DecryptFile(encryptedData, file.EncryptionIV, shares, int(file.Threshold))
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Failed to decrypt file: %v", err),
		})
		return nil, err
	}

	// Handle decompression if needed
	if file.IsCompressed {
		return c.compressionService.Decompress(decryptedData)
	}

	return decryptedData, nil
}

func (c *DownloadFileController) getKeyShares(ctx *gin.Context, file *models.File) ([]services.KeyShare, error) {
	fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
	if err != nil {
		log.Printf("Failed to retrieve key fragments: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve key fragments",
		})
		return nil, err
	}

	if uint(len(fragments)) < file.ShareCount {
		log.Printf("Insufficient key fragments: found %d, expected %d", len(fragments), file.ShareCount)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "File key fragments are incomplete",
		})
		return nil, fmt.Errorf("insufficient key fragments")
	}

	shares := make([]services.KeyShare, len(fragments))
	for i, fragment := range fragments {
		shares[i] = services.KeyShare{
			Index: fragment.FragmentIndex,
			Value: string(fragment.EncryptedFragment),
		}
	}

	return shares, nil
}

func (c *DownloadFileController) logDownloadActivity(user *models.User, file *models.File, ipAddress string) {
	activityDetail := "File downloaded successfully"
	if file.IsCompressed {
		activityDetail = fmt.Sprintf("Compressed file downloaded (%.2f%% of original size)", file.CompressionRatio*100)
	}

	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       user.ID,
		ActivityType: "download",
		FileID:       &file.ID,
		IPAddress:    ipAddress,
		Status:       "success",
		Details:      activityDetail,
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}
}

func (c *DownloadFileController) sendFileResponse(ctx *gin.Context, file *models.File, data []byte) {
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(data)))

	log.Printf("File download successful: %s with threshold %d", file.Name, file.Threshold)
	ctx.Data(http.StatusOK, file.MimeType, data)
}
