package EndUser

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArchiveFileController struct {
	fileModel *models.FileModel
}

func NewArchiveFileController(fileModel *models.FileModel) *ArchiveFileController {
	return &ArchiveFileController{
		fileModel: fileModel,
	}
}

func (c *ArchiveFileController) Archive(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}

	fileID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid file ID",
		})
		return
	}

	err = c.fileModel.ArchiveFile(uint(fileID), userID, ctx.ClientIP())
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "file not found or already archived" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File archived successfully",
	})
}