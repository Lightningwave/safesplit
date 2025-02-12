package EndUser

import (
    "fmt"
    "log"
    "net/http"
    "safesplit/models"
    "safesplit/services"

    "github.com/gin-gonic/gin"
)

type TwoFactorController struct {
    userModel        *models.UserModel
    twoFactorService *services.TwoFactorAuthService
}

func NewTwoFactorController(userModel *models.UserModel, twoFactorService *services.TwoFactorAuthService) *TwoFactorController {
    return &TwoFactorController{
        userModel:        userModel,
        twoFactorService: twoFactorService,
    }
}

func (c *TwoFactorController) GetTwoFactorStatus(ctx *gin.Context) {
    userID, exists := ctx.Get("user_id")
    if !exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
        return
    }

    uid, ok := userID.(uint)
    if !ok {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
        return
    }

    user, err := c.userModel.FindByID(uid)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "two_factor_enabled": user.TwoFactorEnabled,
    })
}

func (c *TwoFactorController) InitiateEnable2FA(ctx *gin.Context) {
    userID, exists := ctx.Get("user_id")
    if !exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
        return
    }

    uid, ok := userID.(uint)
    if !ok {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
        return
    }

    // Get user's email
    user, err := c.userModel.FindByID(uid)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Verify 2FA is not already enabled
    if user.TwoFactorEnabled {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "2FA is already enabled"})
        return
    }

    log.Printf("Initiating 2FA enable for user ID: %d", uid)

    // Send verification code
    if err := c.twoFactorService.SendTwoFactorToken(uid, user.Email); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send verification code: %v", err)})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "Verification code sent to your email"})
}

func (c *TwoFactorController) VerifyAndEnable2FA(ctx *gin.Context) {
    userID, exists := ctx.Get("user_id")
    if !exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
        return
    }

    uid, ok := userID.(uint)
    if !ok {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
        return
    }

    var req struct {
        Code string `json:"code" binding:"required"`
    }

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    // Verify 2FA is not already enabled
    user, err := c.userModel.FindByID(uid)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    if user.TwoFactorEnabled {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "2FA is already enabled"})
        return
    }

    // Verify the code
    if err := c.twoFactorService.VerifyToken(uid, req.Code); err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired verification code"})
        return
    }

    log.Printf("Enabling 2FA for user ID: %d after verification", uid)

    // Enable 2FA
    if err := c.userModel.EnableEmailTwoFactor(uid); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to enable 2FA: %v", err)})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

func (c *TwoFactorController) InitiateDisable2FA(ctx *gin.Context) {
    userID, exists := ctx.Get("user_id")
    if !exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
        return
    }

    uid, ok := userID.(uint)
    if !ok {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
        return
    }

    // Get user's email
    user, err := c.userModel.FindByID(uid)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Verify 2FA is enabled
    if !user.TwoFactorEnabled {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "2FA is not enabled"})
        return
    }

    log.Printf("Initiating 2FA disable for user ID: %d", uid)

    // Send verification code
    if err := c.twoFactorService.SendTwoFactorToken(uid, user.Email); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send verification code: %v", err)})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "Verification code sent to your email"})
}

func (c *TwoFactorController) VerifyAndDisable2FA(ctx *gin.Context) {
    userID, exists := ctx.Get("user_id")
    if !exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
        return
    }

    uid, ok := userID.(uint)
    if !ok {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
        return
    }

    var req struct {
        Code string `json:"code" binding:"required"`
    }

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    // Verify 2FA is still enabled
    user, err := c.userModel.FindByID(uid)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    if !user.TwoFactorEnabled {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "2FA is not enabled"})
        return
    }

    // Verify the code
    if err := c.twoFactorService.VerifyToken(uid, req.Code); err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired verification code"})
        return
    }

    log.Printf("Disabling 2FA for user ID: %d after verification", uid)

    // Disable 2FA
    if err := c.userModel.DisableEmailTwoFactor(uid); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to disable 2FA: %v", err)})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}