package EndUser

import (
	"errors"
	"fmt"
	"net/http"
	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentController struct {
	billingModel *models.BillingModel
	payment      *services.PaymentService
}

func NewPaymentController(billingModel *models.BillingModel) *PaymentController {
	paymentService, err := services.NewPaymentService()
	if err != nil {
		panic("failed to create PaymentService: " + err.Error())
	}
	return &PaymentController{
		billingModel: billingModel,
		payment:      paymentService,
	}
}

func (c *PaymentController) ProcessPayment(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Processing payment for user: %d\n", userID)

	var paymentReq services.PaymentRequest
	if err := ctx.ShouldBindJSON(&paymentReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment request format"})
		return
	}
	fmt.Printf("Payment request received: %+v\n", paymentReq)

	if err := validatePaymentRequest(&paymentReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle billing profile
	existingProfile, err := c.billingModel.GetUserBillingProfile(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check billing profile"})
		return
	}

	var profile *models.BillingProfile
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new profile
		profile = &models.BillingProfile{
			UserID:               userID,
			BillingEmail:         paymentReq.BillingEmail,
			BillingName:          paymentReq.BillingName,
			BillingAddress:       paymentReq.BillingAddress,
			CountryCode:          paymentReq.CountryCode,
			DefaultPaymentMethod: models.PaymentMethodCreditCard,
			BillingCycle:         paymentReq.BillingCycle,
			Currency:             "USD",
			BillingStatus:        models.BillingStatusPending,
		}
		fmt.Printf("Creating new billing profile: %+v\n", profile)
		if err := c.billingModel.CreateBillingProfile(profile); err != nil {
			fmt.Printf("Failed to create billing profile: %v\n", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create billing profile"})
			return
		}
	} else {
		// Update existing profile
		existingProfile.BillingCycle = paymentReq.BillingCycle
		existingProfile.BillingName = paymentReq.BillingName
		existingProfile.BillingEmail = paymentReq.BillingEmail
		existingProfile.BillingAddress = paymentReq.BillingAddress
		existingProfile.CountryCode = paymentReq.CountryCode
		profile = existingProfile

		fmt.Printf("Updating billing profile: %+v\n", profile)
		if err := c.billingModel.UpdateBillingProfile(profile); err != nil {
			fmt.Printf("Failed to update billing profile: %v\n", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update billing profile"})
			return
		}
	}

	// Process payment
	paymentReq.Type = "sale"
	paymentReq.Amount = determineSubscriptionAmount(profile.BillingCycle)
	paymentReq.Currency = profile.Currency

	fmt.Printf("Processing payment with Braintree: %+v\n", paymentReq)
	tx, err := c.payment.ProcessPayment(ctx.Request.Context(), &paymentReq)
	if err != nil {
		fmt.Printf("Payment processing failed: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payment processing failed: " + err.Error()})
		return
	}

	// Update subscription status
	if err := c.billingModel.UpdateSubscriptionStatus(userID, "premium"); err != nil {
		fmt.Printf("Failed to update subscription status: %v\n", err)
		ctx.JSON(http.StatusOK, gin.H{
			"warning":        "Payment successful but subscription update failed. Please contact support.",
			"transaction_id": tx.Id,
			"status":         tx.Status,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.Id,
		"status":         tx.Status,
		"message":        "Payment processed successfully",
	})
}

func (c *PaymentController) GetPaymentStatus(ctx *gin.Context) {
	paymentID := ctx.Query("payment_id")
	if paymentID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	tx, err := c.payment.GetPaymentStatus(ctx.Request.Context(), paymentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment status"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.Id,
		"status":         tx.Status,
		"amount":         tx.Amount.String(),
		"created_at":     tx.CreatedAt,
	})
}

func getUserIDFromContext(ctx *gin.Context) (uint, error) {
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		return 0, errors.New("unauthorized: user ID not found")
	}

	switch v := userIDInterface.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case float64:
		return uint(v), nil
	default:
		return 0, errors.New("invalid user ID format")
	}
}

func validatePaymentRequest(req *services.PaymentRequest) error {
	if req.CardNumber == "" || req.CVV == "" || req.CardHolder == "" {
		return errors.New("missing required payment information")
	}
	if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
		return errors.New("invalid expiry month")
	}
	if req.ExpiryYear < 2024 {
		return errors.New("card has expired")
	}
	if req.BillingCycle != models.BillingCycleMonthly && req.BillingCycle != models.BillingCycleYearly {
		return errors.New("invalid billing cycle")
	}
	return nil
}

func determineSubscriptionAmount(billingCycle string) float64 {
	if billingCycle == models.BillingCycleYearly {
		return 89.99
	}
	return 8.99
}
