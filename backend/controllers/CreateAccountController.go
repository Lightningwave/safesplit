package controllers

import (
	"errors"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type CreateAccountController struct {
	userModel            *models.UserModel
	passwordHistoryModel *models.PasswordHistoryModel
}

func NewCreateAccountController(userModel *models.UserModel, passwordHistoryModel *models.PasswordHistoryModel) *CreateAccountController {
	return &CreateAccountController{
		userModel:            userModel,
		passwordHistoryModel: passwordHistoryModel,
	}
}

type CreateAccountRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type CreateAccountResponse struct {
	User    *models.User `json:"user"`
	Message string       `json:"message"`
}

// CreateAccount handles user registration
func (c *CreateAccountController) CreateAccount(ctx *gin.Context) {
	var req CreateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate user input
	if err := c.validateUserInput(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user object
	user := &models.User{
		Username:           req.Username,
		Email:              req.Email,
		Password:           req.Password,
		Role:               models.RoleEndUser,
		SubscriptionStatus: models.SubscriptionStatusFree,
		StorageQuota:       models.DefaultStorageQuota,
	}

	// Create the user in the database
	createdUser, err := c.userModel.Create(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear sensitive data
	createdUser.Password = ""

	response := CreateAccountResponse{
		User:    createdUser,
		Message: "Account created successfully",
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": response,
	})
}

// validateUserInput contains business logic for validating user input
func (c *CreateAccountController) validateUserInput(req *CreateAccountRequest) error {
	// Password complexity requirements
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasNumber := false
	for _, char := range req.Password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		}
		if char >= 'a' && char <= 'z' {
			hasLower = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	// Username requirements
	if len(req.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	return nil
}
