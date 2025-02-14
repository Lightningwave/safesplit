package controllers

import (
	"fmt"
	"net/http"
	"safesplit/config"
	"safesplit/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginController struct {
	userModel      *models.UserModel
	billingModel   *models.BillingModel
	activityLogger *models.ActivityLogModel
}

type LoginRequest struct {
	Email         string `json:"email" binding:"required"`
	Password      string `json:"password" binding:"required"`
	TwoFactorCode string `json:"two_factor_code"`
}

type UserResponse struct {
	User           *models.User           `json:"user"`
	BillingProfile *models.BillingProfile `json:"billing_profile,omitempty"`
}

func NewLoginController(userModel *models.UserModel, billingModel *models.BillingModel, activityLogger *models.ActivityLogModel) *LoginController {
	return &LoginController{
		userModel:      userModel,
		billingModel:   billingModel,
		activityLogger: activityLogger,
	}
}

func (c *LoginController) Login(ctx *gin.Context) {
	var loginReq LoginRequest
	if err := ctx.ShouldBindJSON(&loginReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userModel.Authenticate(loginReq.Email, loginReq.Password)
	if err != nil {
		var lockedUser *models.User
		lockedUser, _ = c.userModel.FindByEmail(loginReq.Email)

		if lockedUser != nil {
			c.activityLogger.LogActivity(&models.ActivityLog{
				UserID:       lockedUser.ID,
				ActivityType: "login",
				IPAddress:    ctx.ClientIP(),
				Status:       "failure",
				ErrorMessage: err.Error(),
				Details:      fmt.Sprintf("Failed login attempt from IP: %s", ctx.ClientIP()),
				CreatedAt:    time.Now(),
			})
		}

		if lockedUser != nil && lockedUser.AccountLockedUntil != nil && lockedUser.AccountLockedUntil.After(time.Now()) {
			remainingTime := int(lockedUser.AccountLockedUntil.Sub(time.Now()).Minutes())
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error":             fmt.Sprintf("Account locked for %d minutes", remainingTime),
				"status":            "locked",
				"locked_until":      lockedUser.AccountLockedUntil,
				"remaining_minutes": remainingTime,
			})
			return
		}

		if strings.Contains(err.Error(), "attempts remaining") {
			parts := strings.Split(err.Error(), " ")
			for i, part := range parts {
				if part == "remaining" && i > 0 {
					attempts := parts[i-1]
					ctx.JSON(http.StatusUnauthorized, gin.H{
						"error":              err.Error(),
						"status":             "failed",
						"remaining_attempts": attempts,
					})
					return
				}
			}
		}

		if lockedUser != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":  err.Error(),
				"status": "failed",
			})
			return
		}

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":  "Invalid credentials",
			"status": "failed",
		})
		return
	}

	// Handle 2FA if enabled
	if user.TwoFactorEnabled {
		if loginReq.TwoFactorCode == "" {
			if err := c.userModel.InitiateEmailTwoFactor(user.ID); err != nil {
				c.activityLogger.LogActivity(&models.ActivityLog{
					UserID:       user.ID,
					ActivityType: "login",
					IPAddress:    ctx.ClientIP(),
					Status:       "failure",
					ErrorMessage: "Failed to send 2FA code",
					Details:      fmt.Sprintf("2FA initiation failed from IP: %s", ctx.ClientIP()),
					CreatedAt:    time.Now(),
				})
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to send 2FA code",
				})
				return
			}
			ctx.JSON(http.StatusAccepted, gin.H{
				"message":      "2FA required",
				"requires_2fa": true,
				"user_id":      user.ID,
			})
			return
		}

		if err := c.userModel.VerifyEmailTwoFactor(user.ID, loginReq.TwoFactorCode); err != nil {
			c.activityLogger.LogActivity(&models.ActivityLog{
				UserID:       user.ID,
				ActivityType: "login",
				IPAddress:    ctx.ClientIP(),
				Status:       "failure",
				ErrorMessage: "Invalid 2FA code",
				Details:      fmt.Sprintf("Invalid 2FA code from IP: %s", ctx.ClientIP()),
				CreatedAt:    time.Now(),
			})
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 2FA code"})
			return
		}
	}

	token, err := config.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.activityLogger.LogActivity(&models.ActivityLog{
			UserID:       user.ID,
			ActivityType: "login",
			IPAddress:    ctx.ClientIP(),
			Status:       "failure",
			ErrorMessage: "Error generating token",
			Details:      fmt.Sprintf("Token generation failed from IP: %s", ctx.ClientIP()),
			CreatedAt:    time.Now(),
		})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Get billing profile
	user.Password = ""
	billingProfile, err := c.billingModel.GetUserBillingProfile(user.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.activityLogger.LogActivity(&models.ActivityLog{
			UserID:       user.ID,
			ActivityType: "login",
			IPAddress:    ctx.ClientIP(),
			Status:       "failure",
			ErrorMessage: "Error fetching billing details",
			Details:      fmt.Sprintf("Billing profile fetch failed from IP: %s", ctx.ClientIP()),
			CreatedAt:    time.Now(),
		})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching billing details"})
		return
	}

	// Log successful login
	c.activityLogger.LogActivity(&models.ActivityLog{
		UserID:       user.ID,
		ActivityType: "login",
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      fmt.Sprintf("Successful login from IP: %s", ctx.ClientIP()),
		CreatedAt:    time.Now(),
	})

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"data": UserResponse{
			User:           user,
			BillingProfile: billingProfile,
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

	user.Password = ""
	billingProfile, err := c.billingModel.GetUserBillingProfile(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching billing details"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": UserResponse{
			User:           user,
			BillingProfile: billingProfile,
		},
		"role": user.Role,
	})
}