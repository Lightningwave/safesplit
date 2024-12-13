package EndUser

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type ViewFilesController struct {
	fileModel *models.FileModel
}

func NewViewFilesController(fileModel *models.FileModel) *ViewFilesController {
	return &ViewFilesController{
		fileModel: fileModel,
	}
}

func (c *ViewFilesController) ListUserFiles(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}

	files, err := c.fileModel.ListUserFiles(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve files",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"files": files,
		},
	})
}
