package controllers

import (
	"errors"
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateAccountController struct {
	db *gorm.DB
}

func NewCreateAccountController(db *gorm.DB) *CreateAccountController {
	return &CreateAccountController{db: db}
}

// CreateAccount handles user registration
func (c *CreateAccountController) CreateAccount(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Business logic before creation
	if err := c.validateUserInput(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Interact with the User model to create the user in the database
	if err := user.Create(c.db); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear sensitive data before sending the response
	user.Password = ""
	ctx.JSON(http.StatusCreated, user)
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
