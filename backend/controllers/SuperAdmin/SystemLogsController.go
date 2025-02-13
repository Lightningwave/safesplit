package SuperAdmin

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SystemLogsController struct {
	activityLogModel *models.ActivityLogModel
}

func NewSystemLogsController(activityLogModel *models.ActivityLogModel) *SystemLogsController {
	return &SystemLogsController{
		activityLogModel: activityLogModel,
	}
}

func (c *SystemLogsController) GetSystemLogs(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	filters := make(map[string]interface{})
	if timestamp := ctx.Query("timestamp"); timestamp != "" {
		filters["timestamp"] = timestamp
	}
	if category := ctx.Query("category"); category != "" {
		filters["activity_type"] = category
	}
	if source := ctx.Query("source"); source != "" {
		filters["source"] = source
	}
	if userID := ctx.Query("user"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = uint(id)
		}
	}

	logs, total, err := c.activityLogModel.GetSystemLogs(filters, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system logs"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
