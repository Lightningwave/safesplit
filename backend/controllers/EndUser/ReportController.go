package EndUser

import (
	"fmt"
	"log"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	feedbackModel *models.FeedbackModel
	fileModel     *models.FileModel
}

func NewReportController(feedbackModel *models.FeedbackModel, fileModel *models.FileModel) *ReportController {
	return &ReportController{
		feedbackModel: feedbackModel,
		fileModel:     fileModel,
	}
}

type FileReportRequest struct {
	FileID  uint   `json:"file_id" binding:"required"`
	Message string `json:"message" binding:"required,min=10"`
}

type ShareReportRequest struct {
	ShareLink string `json:"share_link" binding:"required"`
	Message   string `json:"message" binding:"required,min=10"`
}

func (c *ReportController) ReportFile(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	var req FileReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	file, err := c.fileModel.GetFileByID(req.FileID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found",
		})
		return
	}

	log.Printf("Creating suspicious file report - FileID: %d, Reporter: %d", req.FileID, currentUser.ID)

	feedback := &models.Feedback{
		UserID:  currentUser.ID,
		Type:    models.FeedbackTypeSuspiciousActivity,
		Subject: fmt.Sprintf("Suspicious File Report: %s (ID: %d)", file.OriginalName, file.ID),
		Message: req.Message,
		Status:  models.FeedbackStatusPending,
	}

	if err := c.feedbackModel.Create(feedback); err != nil {
		log.Printf("Failed to create suspicious file report: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to submit report",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"message": "Suspicious activity report submitted successfully. Our security team will investigate.",
			"id":      feedback.ID,
		},
	})
}

func (c *ReportController) ReportShare(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	var req ShareReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	log.Printf("Creating suspicious share report - ShareLink: %s, Reporter: %d", req.ShareLink, currentUser.ID)

	feedback := &models.Feedback{
		UserID:  currentUser.ID,
		Type:    models.FeedbackTypeSuspiciousActivity,
		Subject: fmt.Sprintf("Suspicious Share Link Report: %s", req.ShareLink),
		Message: req.Message,
		Status:  models.FeedbackStatusPending,
	}

	if err := c.feedbackModel.Create(feedback); err != nil {
		log.Printf("Failed to create suspicious share report: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to submit report",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"message": "Suspicious share link report submitted successfully. Our security team will investigate.",
			"id":      feedback.ID,
		},
	})
}

func (c *ReportController) GetUserReports(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	feedbacks, err := c.feedbackModel.GetAllByUserAndType(currentUser.ID, models.FeedbackTypeSuspiciousActivity)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to fetch reports",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   feedbacks,
	})
}