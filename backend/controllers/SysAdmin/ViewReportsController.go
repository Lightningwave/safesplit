package SysAdmin

import (
	"fmt"
	"log"
	"net/http"
	"safesplit/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ViewReportsController handles viewing and managing suspicious activity reports
type ViewReportsController struct {
	feedbackModel *models.FeedbackModel
	userModel     *models.UserModel
}

func NewViewReportsController(feedbackModel *models.FeedbackModel, userModel *models.UserModel) *ViewReportsController {
	return &ViewReportsController{
		feedbackModel: feedbackModel,
		userModel:     userModel,
	}
}

// GetAllReports retrieves all suspicious activity reports with pagination
func (c *ViewReportsController) GetAllReports(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	status := ctx.Query("status")

	filters := map[string]interface{}{
		"type": models.FeedbackTypeSuspiciousActivity,
	}
	if status != "" {
		filters["status"] = status
	}

	reports, total, err := c.feedbackModel.GetAll(filters, page, pageSize)
	if err != nil {
		log.Printf("Error fetching reports: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to fetch reports",
		})
		return
	}

	// Enrich reports with reporter information
	enrichedReports := make([]gin.H, len(reports))
	for i, report := range reports {
		reporter, err := c.userModel.FindByID(report.UserID)
		reporterInfo := gin.H{
			"id":       report.UserID,
			"username": "Unknown",
		}
		if err == nil {
			reporterInfo["username"] = reporter.Username
			reporterInfo["email"] = reporter.Email
		}

		enrichedReports[i] = gin.H{
			"id":         report.ID,
			"subject":    report.Subject,
			"message":    report.Message,
			"status":     report.Status,
			"created_at": report.CreatedAt,
			"reporter":   reporterInfo,
			"details":    report.Details,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"reports": enrichedReports,
			"meta": gin.H{
				"total":       total,
				"page":        page,
				"page_size":   pageSize,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}

// GetReportDetails retrieves detailed information about a specific report
func (c *ViewReportsController) GetReportDetails(ctx *gin.Context) {
	reportID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid report ID",
		})
		return
	}

	report, err := c.feedbackModel.GetByID(uint(reportID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Report not found",
		})
		return
	}

	if report.Type != models.FeedbackTypeSuspiciousActivity {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Not a suspicious activity report",
		})
		return
	}

	reporter, err := c.userModel.FindByID(report.UserID)
	if err != nil {
		log.Printf("Error fetching reporter details: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"id":         report.ID,
			"subject":    report.Subject,
			"message":    report.Message,
			"status":     report.Status,
			"created_at": report.CreatedAt,
			"details":    report.Details,
			"reporter": gin.H{
				"id":       reporter.ID,
				"username": reporter.Username,
				"email":    reporter.Email,
			},
		},
	})
}

// UpdateReportStatus updates the status of a suspicious activity report
func (c *ViewReportsController) UpdateReportStatus(ctx *gin.Context) {
	reportID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid report ID",
		})
		return
	}

	var req struct {
		Status  models.FeedbackStatus `json:"status" binding:"required,oneof=pending in_review resolved"`
		Comment string                `json:"comment" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	if err := c.feedbackModel.UpdateStatusWithComment(uint(reportID), req.Status, req.Comment); err != nil {
		log.Printf("Error updating report: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to update report",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Report status updated to %s", req.Status),
	})
}

// GetReportStats retrieves statistics about reports
func (c *ViewReportsController) GetReportStats(ctx *gin.Context) {
	filters := map[string]interface{}{
		"type": models.FeedbackTypeSuspiciousActivity,
	}
	reports, _, err := c.feedbackModel.GetAll(filters, 1, 1000)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to fetch reports",
		})
		return
	}

	// Calculate statistics
	today := time.Now().Truncate(24 * time.Hour)
	totalCount := len(reports)
	todayCount := 0
	pendingCount := 0
	inReviewCount := 0
	resolvedCount := 0

	for _, r := range reports {
		// Count by status
		switch r.Status {
		case models.FeedbackStatusPending:
			pendingCount++
		case models.FeedbackStatusInReview:
			inReviewCount++
		case models.FeedbackStatusResolved:
			resolvedCount++
		}

		// Count today's reports
		if r.CreatedAt.After(today) {
			todayCount++
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"total": gin.H{
				"all":      totalCount,
				"today":    todayCount,
				"pending":  pendingCount,
				"review":   inReviewCount,
				"resolved": resolvedCount,
			},
		},
	})
}