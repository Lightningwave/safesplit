package PremiumUser

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileRecoveryController struct {
	fileModel *models.FileModel
}

func NewFileRecoveryController(fileModel *models.FileModel) *FileRecoveryController {
	return &FileRecoveryController{
		fileModel: fileModel,
	}
}

// RecoverFile handles the recovery of a deleted file for premium users
func (c *FileRecoveryController) RecoverFile(ctx *gin.Context) {
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

	// Check if user is premium
	if !currentUser.IsPremiumUser() {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "File recovery is a premium feature",
		})
		return
	}

	// Parse file ID from request
	fileID, err := strconv.ParseUint(ctx.Param("fileId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid file ID",
		})
		return
	}

	// Recover the file
	err = c.fileModel.RecoverFile(uint(fileID), currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File recovered successfully",
	})
}

// ListRecoverableFiles returns a list of files that can be recovered
func (c *FileRecoveryController) ListRecoverableFiles(ctx *gin.Context) {
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

	// Check if user is premium
	if !currentUser.IsPremiumUser() {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "Viewing recoverable files is a premium feature",
		})
		return
	}

	// Get only deleted files
	files, err := c.fileModel.GetRecoverableFiles(currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve recoverable files",
		})
		return
	}

	// Format response
	var response []gin.H
	for _, file := range files {
		response = append(response, gin.H{
			"id":         file.ID,
			"name":       file.OriginalName,
			"size":       file.Size,
			"deleted_at": file.DeletedAt,
			"mime_type":  file.MimeType,
			"folder_id":  file.FolderID,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}
