package EndUser

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type PasswordResetController struct {
	userModel            *models.UserModel
	passwordHistoryModel *models.PasswordHistoryModel
}

func NewPasswordResetController(userModel *models.UserModel, passwordHistoryModel *models.PasswordHistoryModel) *PasswordResetController {
	return &PasswordResetController{
		userModel:            userModel,
		passwordHistoryModel: passwordHistoryModel,
	}
}

type PasswordResetRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (c *PasswordResetController) ResetPassword(ctx *gin.Context) {
	// Get authenticated user
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	endUser, ok := user.(*models.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	// Bind request body
	var req PasswordResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	// Reset password using model method
	if err := c.userModel.ResetPassword(endUser.ID, req.CurrentPassword, req.NewPassword, c.passwordHistoryModel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "password reset successful",
	})
}
