package SysAdmin

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ViewUserAccountDetailsController struct {
	userModel *models.UserModel
}

func NewViewUserAccountDetailsController(userModel *models.UserModel) *ViewUserAccountDetailsController {
	return &ViewUserAccountDetailsController{
		userModel: userModel,
	}
}

type UserAccountDetailsResponse struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AccountType  string `json:"account_type"`
	ReadAccess   bool   `json:"read_access"`
	WriteAccess  bool   `json:"write_access"`
	Subscription struct {
		Status        string `json:"status"`
		PaymentMethod string `json:"payment_method"`
		NextInvoice   string `json:"next_invoice"`
	} `json:"subscription"`
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

	// Construct response
	response := UserAccountDetailsResponse{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		AccountType: user.SubscriptionStatus,
		ReadAccess:  user.ReadAccess,
		WriteAccess: user.WriteAccess,
	}

	// Set subscription details directly from user model
	response.Subscription.Status = user.SubscriptionStatus
	response.Subscription.PaymentMethod = user.PaymentMethod

	// Format next billing date if available
	if user.NextBillingDate != nil {
		response.Subscription.NextInvoice = user.NextBillingDate.Format("January 02, 2006")
	} else {
		response.Subscription.NextInvoice = "Not Available"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_details": response,
	})
}
