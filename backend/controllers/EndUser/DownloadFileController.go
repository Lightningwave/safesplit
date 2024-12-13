package EndUser

import (
	"fmt"
	"net/http"
	"os"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DownloadFileController struct {
	fileModel *models.FileModel
}

func NewDownloadFileController(fileModel *models.FileModel) *DownloadFileController {
	return &DownloadFileController{
		fileModel: fileModel,
	}
}

func (c *DownloadFileController) Download(ctx *gin.Context) {
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

	file, err := c.fileModel.GetFileForDownload(uint(fileID), userID)
	if err != nil {
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

	// Check if file exists on disk
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found on server",
		})
		return
	}

	// Set response headers for download
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
	ctx.Header("Content-Type", file.MimeType)

	ctx.File(file.FilePath)
}
