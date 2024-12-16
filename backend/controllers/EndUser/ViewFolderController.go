package EndUser

import (
	"log"
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ViewFolderController struct {
	folderModel *models.FolderModel
}

func NewViewFolderController(folderModel *models.FolderModel) *ViewFolderController {
	return &ViewFolderController{
		folderModel: folderModel,
	}
}

// ListFolders returns all root folders for the user
func (c *ViewFolderController) ListFolders(ctx *gin.Context) {
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

	folders, err := c.folderModel.GetUserFolders(currentUser.ID)
	if err != nil {
		log.Printf("Error fetching user folders: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to fetch folders",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"folders": folders,
		},
	})
}

// GetFolderContents returns the contents of a specific folder
func (c *ViewFolderController) GetFolderContents(ctx *gin.Context) {
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

	folder, err := c.folderModel.GetFolderContents(uint(folderID), currentUser.ID)
	if err != nil {
		log.Printf("Error fetching folder contents: %v", err)
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

	// Get folder path
	path, err := c.folderModel.GetFolderPath(uint(folderID))
	if err != nil {
		log.Printf("Error getting folder path: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"folder": folder,
			"path":   path,
		},
	})
}
