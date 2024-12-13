package EndUser

import (
	"fmt"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ViewFilesController struct {
	fileModel *models.FileModel
	db        *gorm.DB
}

func NewViewFilesController(db *gorm.DB, fileModel *models.FileModel) *ViewFilesController {
	return &ViewFilesController{
		db:        db,
		fileModel: fileModel,
	}
}

func (c *ViewFilesController) ListUserFiles(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := ctx.GetUint("user_id")
	fmt.Println("Retrieved user_id from context:", userID) // Debug log

	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}

	// Query all non-deleted files for the user
	var files []models.File
	err := c.db.Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("created_at DESC").
		Find(&files).Error

	if err != nil {
		fmt.Println("Database query error:", err) // Debug log
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve files",
		})
		return
	}

	// Return the files
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"files": files,
		},
	})
}
