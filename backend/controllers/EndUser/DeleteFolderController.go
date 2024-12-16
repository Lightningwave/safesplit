package EndUser

import (
	"log"
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteFolderController struct {
	folderModel      *models.FolderModel
	activityLogModel *models.ActivityLogModel
}

func NewDeleteFolderController(
	folderModel *models.FolderModel,
	activityLogModel *models.ActivityLogModel,
) *DeleteFolderController {
	return &DeleteFolderController{
		folderModel:      folderModel,
		activityLogModel: activityLogModel,
	}
}

func (c *DeleteFolderController) Delete(ctx *gin.Context) {
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

	folderID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid folder ID",
		})
		return
	}

	// Get folder path before deletion for logging
	path, err := c.folderModel.GetFolderPath(uint(folderID))
	folderPath := "Unknown folder"
	if err == nil && len(path) > 0 {
		folderPath = path[len(path)-1].Name
		if len(path) > 1 {
			folderPath = path[0].Name + " > " + folderPath
		}
	}

	// Delete the folder
	err = c.folderModel.DeleteFolder(uint(folderID), currentUser.ID)
	if err != nil {
		log.Printf("Error deleting folder: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "folder not found" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// Log activity
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       currentUser.ID,
		ActivityType: "delete",
		Details:      "Deleted folder: " + folderPath,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Folder deleted successfully",
		"data": gin.H{
			"folder_path": folderPath,
		},
	})
}
