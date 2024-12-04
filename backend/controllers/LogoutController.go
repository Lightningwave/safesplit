package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LogoutController struct {
	db *gorm.DB
}

func NewLogoutController(db *gorm.DB) *LogoutController {
	return &LogoutController{db: db}
}

func (c *LogoutController) Logout(ctx *gin.Context) {
	// Since we're using JWT, just return success
	// Frontend will handle token removal
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}
