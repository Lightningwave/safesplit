package SysAdmin

import (
	"fmt"
	"log"
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ViewFeedbacksController handles system admin feedback management
type ViewFeedbacksController struct {
	feedbackModel *models.FeedbackModel
}

// NewViewFeedbacksController creates a new ViewFeedbacksController instance
func NewViewFeedbacksController(feedbackModel *models.FeedbackModel) *ViewFeedbacksController {
	return &ViewFeedbacksController{
		feedbackModel: feedbackModel,
	}
}

// ListFeedbacksRequest represents the request structure for listing feedbacks
type ListFeedbacksRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Status   string `form:"status"`
}

// GetAllFeedbacks retrieves all user feedbacks with filtering and pagination
func (c *ViewFeedbacksController) GetAllFeedbacks(ctx *gin.Context) {
	var req ListFeedbacksRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request parameters",
		})
		return
	}

	filters := make(map[string]interface{})
	filters["type"] = models.FeedbackTypeFeedback
	if req.Status != "" {
		filters["status"] = req.Status
	}

	log.Printf("Fetching feedbacks with filters: %v, page: %d, size: %d", 
		filters, req.Page, req.PageSize)

	feedbacks, total, err := c.feedbackModel.GetAll(filters, req.Page, req.PageSize)
	if err != nil {
		log.Printf("Error fetching feedbacks: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to fetch feedbacks",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"feedbacks": feedbacks,
			"meta": gin.H{
				"total":      total,
				"page":       req.Page,
				"page_size":  req.PageSize,
				"total_pages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
			},
		},
	})
}

// GetFeedback retrieves a single feedback entry by ID
func (c *ViewFeedbacksController) GetFeedback(ctx *gin.Context) {
	feedbackID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid feedback ID",
		})
		return
	}

	feedback, err := c.feedbackModel.GetByID(uint(feedbackID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Feedback not found",
		})
		return
	}

	// Only return if it's general feedback
	if feedback.Type != models.FeedbackTypeFeedback {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Not a general feedback entry",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   feedback,
	})
}

// UpdateFeedbackStatus updates the status of a feedback entry
func (c *ViewFeedbacksController) UpdateFeedbackStatus(ctx *gin.Context) {
	feedbackID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid feedback ID",
		})
		return
	}

	var req struct {
		Status models.FeedbackStatus `json:"status" binding:"required,oneof=pending in_review resolved"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid status",
		})
		return
	}

	log.Printf("Updating feedback status - ID: %d, New Status: %s", feedbackID, req.Status)

	if err := c.feedbackModel.UpdateStatus(uint(feedbackID), req.Status); err != nil {
		log.Printf("Error updating feedback status: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to update feedback status",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Feedback status updated to %s", req.Status),
	})
}

// GetFeedbackStats retrieves statistics about user feedback
func (c *ViewFeedbacksController) GetFeedbackStats(ctx *gin.Context) {
	pendingCount, err := c.feedbackModel.GetPendingCount()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to get feedback statistics",
		})
		return
	}

	feedbacks, _, err := c.feedbackModel.GetAll(map[string]interface{}{
		"type": models.FeedbackTypeFeedback,
	}, 1, 1000)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to get feedback data",
		})
		return
	}

	// Calculate statistics
	totalCount := len(feedbacks)
	resolvedCount := 0
	inReviewCount := 0
	for _, f := range feedbacks {
		switch f.Status {
		case models.FeedbackStatusResolved:
			resolvedCount++
		case models.FeedbackStatusInReview:
			inReviewCount++
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"total":    totalCount,
			"pending":  pendingCount,
			"review":   inReviewCount,
			"resolved": resolvedCount,
		},
	})
}