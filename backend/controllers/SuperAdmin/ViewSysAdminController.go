package SuperAdmin

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type ViewSysAdminController struct {
	userModel *models.UserModel
}

func NewViewSysAdminController(userModel *models.UserModel) *ViewSysAdminController {
	return &ViewSysAdminController{
		userModel: userModel,
	}
}

func (c *ViewSysAdminController) ListSysAdmins(ctx *gin.Context) {
	superAdmin, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	sysAdmins, err := c.userModel.GetSysAdmins(superAdmin.(*models.User))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sysAdmins": sysAdmins})
}
