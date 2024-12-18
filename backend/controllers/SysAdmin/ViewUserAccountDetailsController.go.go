package SysAdmin

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ViewUserAccountDetailsController struct {
	userModel    *models.UserModel
	billingModel *models.BillingModel
}

func NewViewUserAccountDetailsController(userModel *models.UserModel, billingModel *models.BillingModel) *ViewUserAccountDetailsController {
	return &ViewUserAccountDetailsController{
		userModel:    userModel,
		billingModel: billingModel,
	}
}

type UserAccountDetailsResponse struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AccountType  string `json:"account_type"`
	ReadAccess   bool   `json:"read_access"`
	WriteAccess  bool   `json:"write_access"`
	IsActive     bool   `json:"is_active"`
	LastLogin    string `json:"last_login,omitempty"`
	Subscription struct {
		Status          string `json:"status"`
		BillingName     string `json:"billing_name,omitempty"`
		BillingEmail    string `json:"billing_email,omitempty"`
		PaymentMethod   string `json:"payment_method"`
		BillingCycle    string `json:"billing_cycle,omitempty"`
		BillingStatus   string `json:"billing_status"`
		NextInvoiceDate string `json:"next_invoice_date,omitempty"`
	} `json:"subscription"`
	Storage struct {
		QuotaUsed  int64 `json:"quota_used"`
		QuotaTotal int64 `json:"quota_total"`
	} `json:"storage"`
}

func (c *ViewUserAccountDetailsController) GetUserAccountDetails(ctx *gin.Context) {
	// Authenticate admin user
	admin, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	sysAdmin, ok := admin.(*models.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	// Verify admin permissions
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Get user ID from URL parameter
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user details
	user, err := c.userModel.FindByID(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get billing profile
	billingProfile, err := c.billingModel.GetUserBillingProfile(uint(userID))
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching billing details"})
		return
	}

	// Construct response
	response := UserAccountDetailsResponse{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		AccountType: user.Role,
		ReadAccess:  user.ReadAccess,
		WriteAccess: user.WriteAccess,
		IsActive:    user.IsActive,
	}

	// Format last login time if available
	if user.LastLogin != nil {
		response.LastLogin = user.LastLogin.Format("2006-01-02 15:04:05")
	}

	// Set storage information
	response.Storage.QuotaUsed = user.StorageUsed
	response.Storage.QuotaTotal = user.StorageQuota

	// Set subscription details
	response.Subscription.Status = user.SubscriptionStatus

	// Add billing profile details if available
	if billingProfile != nil {
		response.Subscription.BillingName = billingProfile.BillingName
		response.Subscription.BillingEmail = billingProfile.BillingEmail
		response.Subscription.PaymentMethod = billingProfile.DefaultPaymentMethod
		response.Subscription.BillingCycle = billingProfile.BillingCycle
		response.Subscription.BillingStatus = billingProfile.BillingStatus

		if billingProfile.NextBillingDate != nil {
			response.Subscription.NextInvoiceDate = billingProfile.NextBillingDate.Format("January 02, 2006")
		}
	} else {
		response.Subscription.PaymentMethod = "none"
		response.Subscription.BillingStatus = "inactive"
		response.Subscription.NextInvoiceDate = "Not Available"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_details": response,
	})
}
