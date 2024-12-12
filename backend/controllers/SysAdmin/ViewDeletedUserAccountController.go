package SysAdmin

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ViewDeletedUserAccountController struct {
	userModel *models.UserModel
}

func NewViewDeletedUserAccountController(userModel *models.UserModel) *ViewDeletedUserAccountController {
	return &ViewDeletedUserAccountController{
		userModel: userModel,
	}
}

// GetDeletedUsers retrieves all deleted user accounts
func (c *ViewDeletedUserAccountController) GetDeletedUsers(ctx *gin.Context) {
	// Get the authenticated admin from context
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

	// Fetch deleted users
	users, err := c.userModel.GetDeletedUsers(sysAdmin)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

// RestoreUser restores a deleted user account
func (c *ViewDeletedUserAccountController) RestoreUser(ctx *gin.Context) {
	// Get the authenticated admin from context
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

	// Parse user ID from URL parameter
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Attempt to restore the user
	err = c.userModel.RestoreUser(sysAdmin, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user account restored successfully",
	})
}
