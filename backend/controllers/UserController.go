package controllers

import (
	"net/http"
	"safesplit/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserManagementController struct {
	db *gorm.DB
}

func NewUserManagementController(db *gorm.DB) *UserManagementController {
	return &UserManagementController{db: db}
}

// Create a new user (registration)
func (c *UserManagementController) Create(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user with default role
	if err := user.Create(c.db); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear password before sending response
	user.Password = ""
	ctx.JSON(http.StatusCreated, user)
}

// GetMe retrieves the current user's profile
func (c *UserManagementController) GetMe(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var user models.User
	if err := user.FindByID(c.db, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	ctx.JSON(http.StatusOK, user)
}

// ListUsers returns all users (for admin)
func (c *UserManagementController) ListUsers(ctx *gin.Context) {
	var users []models.User
	if err := c.db.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear passwords
	for i := range users {
		users[i].Password = ""
	}

	ctx.JSON(http.StatusOK, users)
}

// CreateSysAdmin creates a system admin user
func (c *UserManagementController) CreateSysAdmin(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.Role = models.RoleSysAdmin
	if err := user.Create(c.db); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	ctx.JSON(http.StatusCreated, user)
}

// DeleteUser deletes a user by ID
func (c *UserManagementController) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("id")

	if err := c.db.Delete(&models.User{}, userID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// AdminDashboard retrieves admin dashboard data
func (c *UserManagementController) AdminDashboard(ctx *gin.Context) {
	var stats struct {
		TotalUsers   int64 `json:"total_users"`
		RegularUsers int64 `json:"regular_users"`
		PremiumUsers int64 `json:"premium_users"`
	}

	// Get total users
	c.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// Get regular users
	c.db.Model(&models.User{}).Where("role = ?", models.RoleEndUser).Count(&stats.RegularUsers)

	// Get premium users
	c.db.Model(&models.User{}).Where("role = ?", models.RolePremiumUser).Count(&stats.PremiumUsers)

	ctx.JSON(http.StatusOK, stats)
}
