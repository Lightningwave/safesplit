package EndUser

import (
	"log"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

// FeedbackController handles user feedback submissions
type FeedbackController struct {
	feedbackModel *models.FeedbackModel
}

// NewFeedbackController creates a new FeedbackController instance
func NewFeedbackController(feedbackModel *models.FeedbackModel) *FeedbackController {
	return &FeedbackController{
		feedbackModel: feedbackModel,
	}
}

// FeedbackRequest represents the request structure for submitting feedback
type FeedbackRequest struct {
	Category string `json:"category" binding:"required,oneof=feature_request bug_report general_feedback improvement suggestion"`
	Subject  string `json:"subject" binding:"required,min=5,max=255"`
	Message  string `json:"message" binding:"required,min=10"`
}

// SubmitFeedback handles user feedback submission
func (c *FeedbackController) SubmitFeedback(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	var req FeedbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	log.Printf("Creating user feedback - User: %d, Category: %s", currentUser.ID, req.Category)

	feedback := &models.Feedback{
		UserID:  currentUser.ID,
		Type:    models.FeedbackTypeFeedback,
		Subject: req.Subject,
		Message: req.Message,
		Status:  models.FeedbackStatusPending,
		Details: req.Category, 
	}

	if err := c.feedbackModel.Create(feedback); err != nil {
		log.Printf("Failed to create feedback: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to submit feedback",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"message": "Thank you for your feedback! We value your input for improving SafeSplit.",
			"id":      feedback.ID,
		},
	})
}

// GetUserFeedback retrieves all feedback submitted by the user
func (c *FeedbackController) GetUserFeedback(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	feedbacks, err := c.feedbackModel.GetAllByUserAndType(currentUser.ID, models.FeedbackTypeFeedback)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to fetch feedback",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   feedbacks,
	})
}

// GetFeedbackCategories returns available feedback categories
func (c *FeedbackController) GetFeedbackCategories(ctx *gin.Context) {
	categories := []gin.H{
		{
			"id": "feature_request",
			"name": "Feature Request",
			"description": "Suggest a new feature for SafeSplit",
		},
		{
			"id": "bug_report",
			"name": "Bug Report",
			"description": "Report a problem or issue with the application",
		},
		{
			"id": "general_feedback",
			"name": "General Feedback",
			"description": "Share your thoughts about SafeSplit",
		},
		{
			"id": "improvement",
			"name": "Improvement Suggestion",
			"description": "Suggest improvements to existing features",
		},
		{
			"id": "suggestion",
			"name": "Other Suggestion",
			"description": "Any other suggestions for SafeSplit",
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categories,
	})
}