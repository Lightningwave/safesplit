package EndUser

import (
	"net/http"
	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
)

type PasswordResetController struct {
	userModel            *models.UserModel
	passwordHistoryModel *models.PasswordHistoryModel
	keyRotationModel     *models.KeyRotationModel
}

func NewPasswordResetController(
	userModel *models.UserModel,
	passwordHistoryModel *models.PasswordHistoryModel,
	keyRotationModel *models.KeyRotationModel,
) *PasswordResetController {
	return &PasswordResetController{
		userModel:            userModel,
		passwordHistoryModel: passwordHistoryModel,
		keyRotationModel:     keyRotationModel,
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

	// First, verify the current password can decrypt the master key
	currentKEK, err := services.DeriveKeyEncryptionKey(req.CurrentPassword, endUser.MasterKeySalt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process current password"})
		return
	}

	// Try to decrypt the current master key
	masterKey, err := services.DecryptMasterKey(
		endUser.EncryptedMasterKey,
		currentKEK,
		endUser.MasterKeyNonce,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect"})
		return
	}

	// Generate new KEK from new password
	newKEK, err := services.DeriveKeyEncryptionKey(req.NewPassword, endUser.MasterKeySalt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process new password"})
		return
	}

	// Re-encrypt master key with new password-derived key
	newEncryptedKey, err := services.EncryptMasterKey(masterKey, newKEK, endUser.MasterKeyNonce)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to secure master key"})
		return
	}

	// Reset password and update master key
	if err := c.userModel.ResetPassword(
		endUser.ID,
		req.CurrentPassword,
		req.NewPassword,
		newEncryptedKey,
		c.passwordHistoryModel,
	); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "password reset successful",
		"details": "master key has been secured with new password",
	})
}
