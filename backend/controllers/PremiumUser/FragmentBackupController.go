package PremiumUser

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FragmentController struct {
	keyFragmentModel *models.KeyFragmentModel
	fileModel        *models.FileModel
}

func NewFragmentController(
	keyFragmentModel *models.KeyFragmentModel,
	fileModel *models.FileModel,
) *FragmentController {
	return &FragmentController{
		keyFragmentModel: keyFragmentModel,
		fileModel:        fileModel,
	}
}

func (c *FragmentController) GetUserFragments(ctx *gin.Context) {
	// Get authenticated user
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

	// Parse file ID
	fileID, err := strconv.ParseUint(ctx.Param("fileId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid file ID",
		})
		return
	}

	// Verify file access
	file, err := c.fileModel.GetFileForDownload(uint(fileID), currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found or access denied",
		})
		return
	}

	// Get user fragments
	fragments, err := c.keyFragmentModel.GetUserFragmentsForFile(uint(fileID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve fragments",
		})
		return
	}

	// Format response
	response := make([]gin.H, len(fragments))
	for i, fragment := range fragments {
		response[i] = gin.H{
			"index": fragment.FragmentIndex,
			"value": string(fragment.EncryptedFragment),
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"fileId":    fileID,
			"fileName":  file.OriginalName,
			"fragments": response,
		},
	})
}
