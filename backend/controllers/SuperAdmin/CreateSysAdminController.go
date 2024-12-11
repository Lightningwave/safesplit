package SuperAdmin

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type CreateSysAdminController struct {
	userModel *models.UserModel
}

func NewCreateSysAdminController(userModel *models.UserModel) *CreateSysAdminController {
	return &CreateSysAdminController{
		userModel: userModel,
	}
}

type CreateSysAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (c *CreateSysAdminController) CreateSysAdmin(ctx *gin.Context) {
	// Get the authenticated user (super admin) from the context
	creator, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	superAdmin, ok := creator.(*models.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	// Parse and validate request body
	var req CreateSysAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	// Create the new admin user object
	newAdmin := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Attempt to create the sys admin
	createdAdmin, err := c.userModel.CreateSysAdmin(superAdmin, newAdmin)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return the created admin (excluding sensitive fields)
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "system administrator created successfully",
		"admin": gin.H{
			"id":       createdAdmin.ID,
			"username": createdAdmin.Username,
			"email":    createdAdmin.Email,
			"role":     createdAdmin.Role,
		},
	})
}
