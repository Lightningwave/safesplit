package SysAdmin

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type ViewUserAccountController struct {
	userModel *models.UserModel
}

func NewViewUserAccountController(userModel *models.UserModel) *ViewUserAccountController {
	return &ViewUserAccountController{
		userModel: userModel,
	}
}

func (c *ViewUserAccountController) ListUsers(ctx *gin.Context) {
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

	users, err := c.userModel.GetAllUsers(sysAdmin)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}
