package EndUser

import (
    "errors"
    "net/http"
    "time"
    "safesplit/models"
    "github.com/gin-gonic/gin"
)

type SubscriptionController struct {
    billingModel *models.BillingModel
}

type CancellationResponse struct {
    Message        string    `json:"message"`
    RemainingDays  int       `json:"remaining_days"`
    EndDate        time.Time `json:"end_date"`
    DowngradeInfo  string    `json:"downgrade_info"`
}

func NewSubscriptionController(billingModel *models.BillingModel) *SubscriptionController {
    return &SubscriptionController{
        billingModel: billingModel,
    }
}

func (c *SubscriptionController) CancelSubscription(ctx *gin.Context) {
    userID, err := getUserIDFromContext(ctx)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    profile, err := c.billingModel.GetUserBillingProfile(userID)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Billing profile not found"})
        return
    }

    if profile.BillingStatus != models.BillingStatusActive {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "No active subscription found",
            "status": profile.BillingStatus,
        })
        return
    }

    if err := c.billingModel.CancelSubscription(userID); err != nil {
        if errors.Is(err, models.ErrStorageExceedsQuota) {
            ctx.JSON(http.StatusConflict, gin.H{
                "error": "Cannot downgrade: storage usage exceeds free tier quota",
                "current_usage": profile.User.StorageUsed,
                "free_quota": models.DefaultStorageQuota,
            })
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
        return
    }

    remainingDays := getRemainingDays(profile.NextBillingDate)
    response := CancellationResponse{
        Message:       "Subscription cancelled successfully",
        RemainingDays: remainingDays,
        EndDate:       *profile.NextBillingDate,
        DowngradeInfo: "Account will be downgraded to free tier after current billing period",
    }

    ctx.JSON(http.StatusOK, response)
}

func getRemainingDays(endDate *time.Time) int {
    if endDate == nil {
        return 0
    }
    return int(time.Until(*endDate).Hours() / 24)
}