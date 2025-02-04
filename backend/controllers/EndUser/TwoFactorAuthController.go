package EndUser

import (
	"fmt"
	"log"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type TwoFactorController struct {
	userModel *models.UserModel
}

func NewTwoFactorController(userModel *models.UserModel) *TwoFactorController {
	return &TwoFactorController{
		userModel: userModel,
	}
}

func (c *TwoFactorController) EnableEmailTwoFactor(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
		return
	}

	log.Printf("Enabling 2FA for user ID: %d", uid)

	if err := c.userModel.EnableEmailTwoFactor(uid); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to enable 2FA: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

func (c *TwoFactorController) DisableEmailTwoFactor(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
		return
	}

	log.Printf("Disabling 2FA for user ID: %d", uid)

	if err := c.userModel.DisableEmailTwoFactor(uid); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to disable 2FA: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

func (c *TwoFactorController) GetTwoFactorStatus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
		return
	}

	user, err := c.userModel.FindByID(uid)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"two_factor_enabled": user.TwoFactorEnabled,
	})
}
