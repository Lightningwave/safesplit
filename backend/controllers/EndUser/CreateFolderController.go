package EndUser

import (
	"log"
	"net/http"
	"safesplit/models"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateFolderController struct {
	folderModel      *models.FolderModel
	activityLogModel *models.ActivityLogModel
}

func NewCreateFolderController(
	folderModel *models.FolderModel,
	activityLogModel *models.ActivityLogModel,
) *CreateFolderController {
	return &CreateFolderController{
		folderModel:      folderModel,
		activityLogModel: activityLogModel,
	}
}

func (c *CreateFolderController) Create(ctx *gin.Context) {
	log.Printf("Starting folder creation request")

	// Get authenticated user
	user, exists := ctx.Get("user")
	if !exists {
		log.Printf("User authentication failed - user not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		log.Printf("User authentication failed - invalid user type in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Invalid user data",
		})
		return
	}

	// Parse request body
	var req struct {
		Name           string `json:"name" binding:"required"`
		ParentFolderID *uint  `json:"parent_folder_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request data: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	// Create folder object
	newFolder := &models.Folder{
		UserID:         currentUser.ID,
		Name:           req.Name,
		ParentFolderID: req.ParentFolderID,
		IsArchived:     false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create folder
	err := c.folderModel.CreateFolder(newFolder)
	if err != nil {
		log.Printf("Failed to create folder: %v", err)
		status := http.StatusInternalServerError

		switch err.Error() {
		case "folder name is required":
			status = http.StatusBadRequest
		case "invalid parent folder":
			status = http.StatusNotFound
		}

		ctx.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// Get folder path for activity log
	var folderPath string
	if path, err := c.folderModel.GetFolderPath(newFolder.ID); err == nil {
		for i, f := range path {
			if i > 0 {
				folderPath += " > "
			}
			folderPath += f.Name
		}
	}

	// Log activity
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       currentUser.ID,
		ActivityType: "create",
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      "Created folder: " + folderPath,
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	// Return response with folder path
	var response gin.H = gin.H{
		"id":               newFolder.ID,
		"name":             newFolder.Name,
		"parent_folder_id": newFolder.ParentFolderID,
		"created_at":       newFolder.CreatedAt,
	}

	if folderPath != "" {
		response["path"] = folderPath
	}

	log.Printf("Folder created successfully: %s", folderPath)
	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Folder created successfully",
		"data": gin.H{
			"folder": response,
		},
	})
}
