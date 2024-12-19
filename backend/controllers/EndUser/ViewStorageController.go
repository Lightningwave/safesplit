package EndUser

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type ViewStorageController struct {
	fileModel *models.FileModel
	userModel *models.UserModel
}

func NewViewStorageController(fileModel *models.FileModel, userModel *models.UserModel) *ViewStorageController {
	return &ViewStorageController{
		fileModel: fileModel,
		userModel: userModel,
	}
}

// GetStorageInfo returns storage usage information for the current user
func (c *ViewStorageController) GetStorageInfo(ctx *gin.Context) {
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

	// Get storage information using existing method
	used, quota, err := c.fileModel.GetUserStorageInfo(currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve storage information",
		})
		return
	}

	// Get total files count
	fileCount, err := c.fileModel.GetUserFileCount(currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve file count",
		})
		return
	}

	// Calculate percentage used
	var percentageUsed float64
	if quota > 0 {
		percentageUsed = float64(used) / float64(quota) * 100
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"storage_used":    used,
			"storage_quota":   quota,
			"total_files":     fileCount,
			"percentage_used": percentageUsed,
			"is_premium":      currentUser.IsPremiumUser(),
		},
	})
}
