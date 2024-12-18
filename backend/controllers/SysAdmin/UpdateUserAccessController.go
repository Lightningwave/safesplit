package SysAdmin

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UpdateAccountController struct {
	userModel *models.UserModel
}

func NewUpdateAccountController(userModel *models.UserModel) *UpdateAccountController {
	return &UpdateAccountController{
		userModel: userModel,
	}
}

type UpdateAccountRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=255"`
	Email       string `json:"email" binding:"required,email"`
	AccountType string `json:"account_type" binding:"required,oneof=free premium"`
	ReadAccess  bool   `json:"read_access"`
	WriteAccess bool   `json:"write_access"`
}

func (c *UpdateAccountController) UpdateAccount(ctx *gin.Context) {
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

	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req UpdateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	// Create user update object
	userUpdate := &models.User{
		Username:           req.Username,
		Email:              req.Email,
		SubscriptionStatus: req.AccountType,
		ReadAccess:         req.ReadAccess,
		WriteAccess:        req.WriteAccess,
	}

	// Update user account
	if err := c.userModel.UpdateUserAccount(sysAdmin, uint(userID), userUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated user information
	updatedUser, err := c.userModel.FindByID(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated information"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "account updated successfully",
		"data": gin.H{
			"user": updatedUser,
		},
	})
}
