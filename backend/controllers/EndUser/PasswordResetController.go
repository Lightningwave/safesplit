package EndUser

import (
	"log"
	"net/http"
	"safesplit/models"
	"time"

	"github.com/gin-gonic/gin"
)

type PasswordResetController struct {
    userModel            *models.UserModel
    passwordHistoryModel *models.PasswordHistoryModel
    keyFragmentModel     *models.KeyFragmentModel  
    fileModel           *models.FileModel         
}

type PasswordResetRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func NewPasswordResetController(
    userModel *models.UserModel,
    passwordHistoryModel *models.PasswordHistoryModel,
    keyFragmentModel *models.KeyFragmentModel,
    fileModel *models.FileModel,
) *PasswordResetController {
    return &PasswordResetController{
        userModel:            userModel,
        passwordHistoryModel: passwordHistoryModel,
        keyFragmentModel:     keyFragmentModel,
        fileModel:            fileModel,
    }
}

func (c *PasswordResetController) ResetPassword(ctx *gin.Context) {
    startTime := time.Now()
    log.Printf("Starting password reset process")

    user, exists := ctx.Get("user")
    if !exists {
        log.Printf("Password reset failed: No authenticated user found")
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    endUser, ok := user.(*models.User)
    if !ok {
        log.Printf("Password reset failed: Invalid user type in context")
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
        return
    }

    var req PasswordResetRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        log.Printf("Password reset failed: Invalid request data - %v", err)
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
        return
    }

    err := c.userModel.ResetPasswordWithFragments(
        endUser.ID,
        req.CurrentPassword,
        req.NewPassword,
        c.passwordHistoryModel,
        c.keyFragmentModel,
        c.fileModel,
    )

    if err != nil {
        log.Printf("Password reset failed for user %d: %v", endUser.ID, err)
        
        status := http.StatusInternalServerError
        switch err.Error() {
        case "current password is incorrect":
            status = http.StatusBadRequest
        case "user not found":
            status = http.StatusNotFound
        }
        
        ctx.JSON(status, gin.H{"error": err.Error()})
        return
    }

    duration := time.Since(startTime)
    log.Printf("Password reset successful for user %d - Duration: %v", endUser.ID, duration)

    ctx.JSON(http.StatusOK, gin.H{
        "message": "password reset successful",
        "details": "master key and file access keys have been secured with new password",
    })
}