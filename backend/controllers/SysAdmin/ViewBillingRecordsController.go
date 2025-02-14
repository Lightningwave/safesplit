package SysAdmin

import (
    "net/http"
    "safesplit/models"
    "github.com/gin-gonic/gin"
)

type ViewBillingRecordsController struct {
    billingModel *models.BillingModel
}

func NewViewBillingRecordsController(billingModel *models.BillingModel) *ViewBillingRecordsController {
    return &ViewBillingRecordsController{
        billingModel: billingModel,
    }
}

type ListBillingRecordsRequest struct {
    Page     int    `form:"page,default=1" binding:"min=1"`
    PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
    Status   string `form:"status"`
    Cycle    string `form:"cycle"`
}

// GetBillingStats returns subscription statistics
func (c *ViewBillingRecordsController) GetBillingStats(ctx *gin.Context) {
    admin, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    sysAdmin, ok := admin.(*models.User)
    if !ok || (!sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin()) {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access"})
        return
    }

    stats, err := c.billingModel.GetSubscriptionStats()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch billing statistics",
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data":   stats,
    })
}

func (c *ViewBillingRecordsController) GetAllBillingRecords(ctx *gin.Context) {
    admin, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    sysAdmin, ok := admin.(*models.User)
    if !ok || (!sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin()) {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access"})
        return
    }

    var req ListBillingRecordsRequest
    if err := ctx.ShouldBindQuery(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request parameters",
        })
        return
    }

    profiles, totalCount, err := c.billingModel.GetAllBillingRecords(
        req.Status,
        req.Cycle,
        req.Page,
        req.PageSize,
    )
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch billing records",
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data": gin.H{
            "records": profiles,
            "meta": gin.H{
                "total":       totalCount,
                "page":        req.Page,
                "page_size":   req.PageSize,
                "total_pages": (totalCount + int64(req.PageSize) - 1) / int64(req.PageSize),
            },
        },
    })
}

// GetExpiringSubscriptions retrieves subscriptions that will expire soon
func (c *ViewBillingRecordsController) GetExpiringSubscriptions(ctx *gin.Context) {
    admin, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    sysAdmin, ok := admin.(*models.User)
    if !ok || (!sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin()) {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access"})
        return
    }

    // Get subscriptions expiring in next 7 days
    expiringProfiles, err := c.billingModel.GetExpiringSubscriptions(7)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch expiring subscriptions",
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data":   expiringProfiles,
    })
}