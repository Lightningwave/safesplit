package SuperAdmin

import (
	"net/http"
	"safesplit/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteSysAdminController struct {
	userModel *models.UserModel
}

func NewDeleteSysAdminController(userModel *models.UserModel) *DeleteSysAdminController {
	return &DeleteSysAdminController{
		userModel: userModel,
	}
}

func (c *DeleteSysAdminController) DeleteSysAdmin(ctx *gin.Context) {
	// Get the authenticated user (super admin) from the context
	superAdmin, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	creator, ok := superAdmin.(*models.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	// Get the sysadmin ID from the URL parameter
	sysAdminID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid system administrator ID"})
		return
	}

	// Attempt to delete the sysadmin
	if err := c.userModel.DeleteSysAdmin(creator, uint(sysAdminID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "system administrator deleted successfully",
	})
}
