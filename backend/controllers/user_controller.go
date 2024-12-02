package controllers

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{db: db}
}

func (c *UserController) Create(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user in database
	if err := user.Create(c.db); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear password for response
	user.Password = ""
	ctx.JSON(http.StatusCreated, user)
}

func (c *UserController) GetMe(ctx *gin.Context) {
	userID := ctx.GetUint("user_id") // Set by auth middleware

	var user models.User
	if err := user.FindByID(c.db, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	ctx.JSON(http.StatusOK, user)
}
