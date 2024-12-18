package controllers

import (
	"net/http"
	"safesplit/config"
	"safesplit/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginController struct {
	userModel    *models.UserModel
	billingModel *models.BillingModel
}

func NewLoginController(userModel *models.UserModel, billingModel *models.BillingModel) *LoginController {
	return &LoginController{
		userModel:    userModel,
		billingModel: billingModel,
	}
}

type UserResponse struct {
	User           *models.User           `json:"user"`
	BillingProfile *models.BillingProfile `json:"billing_profile,omitempty"`
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

	// Authenticate user
	user, err := c.userModel.Authenticate(loginReq.Email, loginReq.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
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

	// Get billing profile if it exists
	billingProfile, err := c.billingModel.GetUserBillingProfile(user.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching billing details"})
		return
	}

	response := UserResponse{
		User:           user,
		BillingProfile: billingProfile,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"data":  response,
	})
}

func (c *LoginController) GetMe(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	user, err := c.userModel.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Clear sensitive data
	user.Password = ""

	// Get billing profile if it exists
	billingProfile, err := c.billingModel.GetUserBillingProfile(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching billing details"})
		return
	}

	response := UserResponse{
		User:           user,
		BillingProfile: billingProfile,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": response,
		"role": user.Role, // Explicitly include role in response
	})
}
