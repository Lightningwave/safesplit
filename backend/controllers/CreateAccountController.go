package controllers

import (
	"errors"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
)

type CreateAccountController struct {
	userModel *models.UserModel
}

func NewCreateAccountController(userModel *models.UserModel) *CreateAccountController {
	return &CreateAccountController{
		userModel: userModel,
	}
}

// CreateAccount handles user registration
func (c *CreateAccountController) CreateAccount(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate user input before creating the user
	if err := c.validateUserInput(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the user in the database using the UserModel
	createdUser, err := c.userModel.Create(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear sensitive data before sending the response
	createdUser.Password = ""
	ctx.JSON(http.StatusCreated, createdUser)
}

// validateUserInput contains business logic for validating user input
func (c *CreateAccountController) validateUserInput(user *models.User) error {
	// Example validation logic
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	// Add more validation as needed
	return nil
}
