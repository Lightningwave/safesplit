package SysAdmin

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteUserAccountController struct {
	userModel *models.UserModel
}

func NewDeleteUserAccountController(userModel *models.UserModel) *DeleteUserAccountController {
	return &DeleteUserAccountController{
		userModel: userModel,
	}
}

func (c *DeleteUserAccountController) DeleteUser(ctx *gin.Context) {
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

	err = c.userModel.DeleteUser(sysAdmin, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user account deleted successfully",
	})
}
