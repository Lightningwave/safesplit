package EndUser

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type MassDeleteFileController struct {
	fileModel *models.FileModel
}

func NewMassDeleteFileController(fileModel *models.FileModel) *MassDeleteFileController {
	return &MassDeleteFileController{
		fileModel: fileModel,
	}
}

func (c *MassDeleteFileController) Delete(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}

	var request struct {
		FileIDs []uint `json:"file_ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request body",
		})
		return
	}

	results := make(map[uint]string)
	for _, fileID := range request.FileIDs {
		err := c.fileModel.DeleteFile(fileID, userID, ctx.ClientIP())
		if err != nil {
			results[fileID] = err.Error()
		} else {
			results[fileID] = "success"
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"results": results,
	})
}