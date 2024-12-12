package SysAdmin

import (
	"fmt"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type ViewUserStorageController struct {
	userModel *models.UserModel
}

func NewViewUserStorageController(userModel *models.UserModel) *ViewUserStorageController {
	return &ViewUserStorageController{
		userModel: userModel,
	}
}

func (c *ViewUserStorageController) GetStorageStats(ctx *gin.Context) {
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

	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	stats, err := c.userModel.GetStorageStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch storage statistics: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"storage_stats": stats,
	})
}
