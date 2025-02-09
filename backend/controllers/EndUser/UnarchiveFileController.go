package EndUser

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UnarchiveFileController struct {
	fileModel *models.FileModel
}

func NewUnarchiveFileController(fileModel *models.FileModel) *UnarchiveFileController {
	return &UnarchiveFileController{
		fileModel: fileModel,
	}
}

func (c *UnarchiveFileController) Unarchive(ctx *gin.Context) {
	// Get user ID from context
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}

	// Parse file ID from URL parameter
	fileID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid file ID",
		})
		return
	}

	// Call the model method to unarchive the file
	err = c.fileModel.UnarchiveFile(uint(fileID), userID, ctx.ClientIP())
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "file not found or not archived" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// Return success response
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File unarchived successfully",
	})
}