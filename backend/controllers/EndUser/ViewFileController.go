package EndUser

import (
	"net/http"
	"safesplit/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ViewFilesController struct {
	fileModel   *models.FileModel
	folderModel *models.FolderModel
}

func NewViewFilesController(fileModel *models.FileModel, folderModel *models.FolderModel) *ViewFilesController {
	return &ViewFilesController{
		fileModel:   fileModel,
		folderModel: folderModel,
	}
}

type FileResponse struct {
	ID               uint       `json:"id"`
	UserID           uint       `json:"user_id"`
	FolderID         *uint      `json:"folder_id"`
	Name             string     `json:"name"`
	OriginalName     string     `json:"original_name"`
	FilePath         string     `json:"file_path"`
	Size             int64      `json:"size"`
	CompressedSize   int64      `json:"compressed_size"`
	IsCompressed     bool       `json:"is_compressed"`
	CompressionRatio float64    `json:"compression_ratio"`
	MimeType         string     `json:"mime_type"`
	IsArchived       bool       `json:"is_archived"`
	IsDeleted        bool       `json:"is_deleted"`
	DeletedAt        *time.Time `json:"deleted_at"`
	FileHash         string     `json:"file_hash"`
	ShareCount       uint       `json:"share_count"`
	Threshold        uint       `json:"threshold"`
	IsShared         bool       `json:"is_shared"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	FolderName       string     `json:"folder_name"`
}

func (c *ViewFilesController) ListUserFiles(ctx *gin.Context) {
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

	folderIDStr := ctx.Query("folder_id")
	var files []models.File
	var err error

	if folderIDStr != "" {
		// Get files from specific folder
		folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid folder ID",
			})
			return
		}

		// Get folder information
		folder, err := c.folderModel.GetFolderByID(uint(folderID), currentUser.ID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "Folder not found",
			})
			return
		}

		files, err = c.fileModel.ListFolderFiles(currentUser.ID, uint(folderID))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to retrieve folder files",
			})
			return
		}

		// Create responses with folder name
		fileResponses := make([]FileResponse, len(files))
		for i, file := range files {
			fileResponses[i] = FileResponse{
				ID:               file.ID,
				UserID:           file.UserID,
				FolderID:         file.FolderID,
				Name:             file.Name,
				OriginalName:     file.OriginalName,
				FilePath:         file.FilePath,
				Size:             file.Size,
				CompressedSize:   file.CompressedSize,
				IsCompressed:     file.IsCompressed,
				CompressionRatio: file.CompressionRatio,
				MimeType:         file.MimeType,
				IsArchived:       file.IsArchived,
				IsDeleted:        file.IsDeleted,
				DeletedAt:        file.DeletedAt,
				FileHash:         file.FileHash,
				ShareCount:       file.ShareCount,
				Threshold:        file.Threshold,
				IsShared:         file.IsShared,
				CreatedAt:        file.CreatedAt,
				UpdatedAt:        file.UpdatedAt,
				FolderName:       folder.Name,
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"files": fileResponses,
			},
		})
		return
	}

	// Get all user files when no folder is specified
	files, err = c.fileModel.ListAllUserFiles(currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve files",
		})
		return
	}

	// Create response with appropriate folder names
	fileResponses := make([]FileResponse, len(files))
	for i, file := range files {
		folderName := "Root"
		if file.FolderID != nil {
			// Get folder information for each file
			if folder, err := c.folderModel.GetFolderByID(*file.FolderID, currentUser.ID); err == nil {
				folderName = folder.Name
			}
		}

		fileResponses[i] = FileResponse{
			ID:               file.ID,
			UserID:           file.UserID,
			FolderID:         file.FolderID,
			Name:             file.Name,
			OriginalName:     file.OriginalName,
			FilePath:         file.FilePath,
			Size:             file.Size,
			CompressedSize:   file.CompressedSize,
			IsCompressed:     file.IsCompressed,
			CompressionRatio: file.CompressionRatio,
			MimeType:         file.MimeType,
			IsArchived:       file.IsArchived,
			IsDeleted:        file.IsDeleted,
			DeletedAt:        file.DeletedAt,
			FileHash:         file.FileHash,
			ShareCount:       file.ShareCount,
			Threshold:        file.Threshold,
			IsShared:         file.IsShared,
			CreatedAt:        file.CreatedAt,
			UpdatedAt:        file.UpdatedAt,
			FolderName:       folderName,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"files": fileResponses,
		},
	})
}
