package EndUser

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type MassUnarchiveFileController struct {
	fileModel *models.FileModel
}

func NewMassUnarchiveFileController(fileModel *models.FileModel) *MassUnarchiveFileController {
	return &MassUnarchiveFileController{
		fileModel: fileModel,
	}
}

func (c *MassUnarchiveFileController) Unarchive(ctx *gin.Context) {
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
		err := c.fileModel.UnarchiveFile(fileID, userID, ctx.ClientIP())
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