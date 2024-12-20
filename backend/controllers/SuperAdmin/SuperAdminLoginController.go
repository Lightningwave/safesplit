package SuperAdmin

import (
	"net/http"
	"safesplit/config"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type LoginController struct {
	userModel *models.UserModel
}

func NewLoginController(userModel *models.UserModel) *LoginController {
	return &LoginController{
		userModel: userModel,
	}
}

func (c *LoginController) Login(ctx *gin.Context) {
	var loginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&loginReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate super admin
	user, err := c.userModel.AuthenticateSuperAdmin(loginReq.Email, loginReq.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid super admin credentials"})
		return
	}

	// Generate token
	token, err := config.GenerateToken(user.ID, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Clear sensitive data
	user.Password = ""

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"data": gin.H{
			"user": user,
		},
	})
}

func (c *LoginController) GetMe(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	user, err := c.userModel.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify super admin role
	if !user.IsSuperAdmin() {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	user.Password = "" // Clear sensitive data
	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user": user,
		},
		"role": user.Role,
	})
}
