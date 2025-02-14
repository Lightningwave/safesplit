package PremiumUser

import (
    "net/http"
    "safesplit/models"
    "github.com/gin-gonic/gin"
)

type UpdateBillingController struct {
    billingModel *models.BillingModel
}

func NewUpdateBillingController(billingModel *models.BillingModel) *UpdateBillingController {
    return &UpdateBillingController{
        billingModel: billingModel,
    }
}

type UpdateBillingRequest struct {
    BillingName          string `json:"billing_name" binding:"required"`
    BillingEmail         string `json:"billing_email" binding:"required,email"`
    BillingAddress       string `json:"billing_address" binding:"required"`
    CountryCode          string `json:"country_code" binding:"required,len=2"`
    DefaultPaymentMethod string `json:"default_payment_method" binding:"required,oneof=credit_card bank_account paypal"`
    BillingCycle         string `json:"billing_cycle" binding:"required,oneof=monthly yearly"`
    Currency             string `json:"currency" binding:"required,len=3"`
}

// UpdateBillingDetails updates the billing profile for a premium user
func (c *UpdateBillingController) UpdateBillingDetails(ctx *gin.Context) {
    user, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    premiumUser, ok := user.(*models.User)
    if !ok {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
        return
    }

    var req UpdateBillingRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
        return
    }

    // Get existing billing profile or create new one
    profile, err := c.billingModel.GetUserBillingProfile(premiumUser.ID)
    if err != nil {
        // Create new profile if doesn't exist
        profile = &models.BillingProfile{
            UserID:               premiumUser.ID,
            BillingStatus:        models.BillingStatusActive,
        }
    }

    // Update profile with new details
    profile.BillingName = req.BillingName
    profile.BillingEmail = req.BillingEmail
    profile.BillingAddress = req.BillingAddress
    profile.CountryCode = req.CountryCode
    profile.DefaultPaymentMethod = req.DefaultPaymentMethod
    profile.BillingCycle = req.BillingCycle
    profile.Currency = req.Currency

    var updateErr error
    if profile.ID == 0 {
        updateErr = c.billingModel.CreateBillingProfile(profile)
    } else {
        updateErr = c.billingModel.UpdateBillingProfile(profile)
    }

    if updateErr != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to update billing details",
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Billing details updated successfully",
        "data": profile,
    })
}

// GetBillingDetails retrieves the current billing profile
func (c *UpdateBillingController) GetBillingDetails(ctx *gin.Context) {
    user, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    premiumUser, ok := user.(*models.User)
    if !ok {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
        return
    }

    profile, err := c.billingModel.GetUserBillingProfile(premiumUser.ID)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "billing profile not found"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "data": profile,
    })
}