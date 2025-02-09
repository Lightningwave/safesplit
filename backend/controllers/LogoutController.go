package controllers

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type LogoutController struct {
	userModel *models.UserModel
}

func NewLogoutController(userModel *models.UserModel) *LogoutController {
	return &LogoutController{userModel: userModel}
}

func (c *LogoutController) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}
