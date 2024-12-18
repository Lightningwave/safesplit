package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// BillingModel constants
const (
	PaymentMethodCreditCard  = "credit_card"
	PaymentMethodBankAccount = "bank_account"
	PaymentMethodPaypal      = "paypal"

	BillingCycleMonthly = "monthly"
	BillingCycleYearly  = "yearly"

	BillingStatusActive    = "active"
	BillingStatusPending   = "pending"
	BillingStatusFailed    = "failed"
	BillingStatusCancelled = "cancelled"
)

type BillingProfile struct {
	ID                   uint       `json:"id" gorm:"primaryKey"`
	UserID               uint       `json:"user_id"`
	CustomerID           string     `json:"customer_id" gorm:"unique"`
	BillingEmail         string     `json:"billing_email"`
	BillingName          string     `json:"billing_name"`
	BillingAddress       string     `json:"billing_address" gorm:"type:text"`
	CountryCode          string     `json:"country_code" gorm:"type:varchar(2)"`
	DefaultPaymentMethod string     `json:"default_payment_method" gorm:"type:enum('credit_card','bank_account','paypal');default:'credit_card'"`
	BillingCycle         string     `json:"billing_cycle" gorm:"type:enum('monthly','yearly');default:'monthly'"`
	Currency             string     `json:"currency" gorm:"type:varchar(3);default:'USD'"`
	NextBillingDate      *time.Time `json:"next_billing_date"`
	LastBillingDate      *time.Time `json:"last_billing_date"`
	BillingStatus        string     `json:"billing_status" gorm:"type:enum('active','pending','failed','cancelled');default:'pending'"`
	CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	User                 User       `json:"-" gorm:"foreignKey:UserID"`
}

type UserBillingUpdate struct {
	User    *User
	Billing *BillingProfileUpdate
}

type BillingProfileUpdate struct {
	BillingName    string
	BillingEmail   string
	BillingAddress string
	PaymentMethod  string
	BillingCycle   string
	Currency       string
	CountryCode    string
}

type UserBillingInfo struct {
	User           *User           `json:"user"`
	BillingProfile *BillingProfile `json:"billing_profile,omitempty"`
}

type BillingModel struct {
	db        *gorm.DB
	userModel *UserModel
}

func NewBillingModel(db *gorm.DB, userModel *UserModel) *BillingModel {
	return &BillingModel{
		db:        db,
		userModel: userModel,
	}
}

// CreateBillingProfile creates a new billing profile for a user
func (m *BillingModel) CreateBillingProfile(profile *BillingProfile) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		// Check if user already has a billing profile
		var existingProfile BillingProfile
		if err := tx.Where("user_id = ?", profile.UserID).First(&existingProfile).Error; err == nil {
			return errors.New("user already has a billing profile")
		}

		// Create the billing profile
		if err := tx.Create(profile).Error; err != nil {
			return err
		}

		return nil
	})
}

// updateBillingProfile updates or creates a billing profile
func (m *BillingModel) updateBillingProfile(tx *gorm.DB, userID uint, update *BillingProfileUpdate) error {
	var profile BillingProfile
	err := tx.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new profile
			profile = BillingProfile{
				UserID:               userID,
				BillingName:          update.BillingName,
				BillingEmail:         update.BillingEmail,
				BillingAddress:       update.BillingAddress,
				DefaultPaymentMethod: update.PaymentMethod,
				BillingCycle:         update.BillingCycle,
				Currency:             update.Currency,
				CountryCode:          update.CountryCode,
				BillingStatus:        BillingStatusPending,
			}
			return tx.Create(&profile).Error
		}
		return err
	}

	// Update existing profile
	updates := map[string]interface{}{}
	if update.BillingName != "" {
		updates["billing_name"] = update.BillingName
	}
	if update.BillingEmail != "" {
		updates["billing_email"] = update.BillingEmail
	}
	if update.BillingAddress != "" {
		updates["billing_address"] = update.BillingAddress
	}
	if update.PaymentMethod != "" {
		updates["default_payment_method"] = update.PaymentMethod
	}
	if update.BillingCycle != "" {
		updates["billing_cycle"] = update.BillingCycle
	}
	if update.Currency != "" {
		updates["currency"] = update.Currency
	}
	if update.CountryCode != "" {
		updates["country_code"] = update.CountryCode
	}

	if len(updates) > 0 {
		return tx.Model(&profile).Updates(updates).Error
	}

	return nil
}

// GetUserBillingProfile retrieves a user's billing profile
func (m *BillingModel) GetUserBillingProfile(userID uint) (*BillingProfile, error) {
	var profile BillingProfile
	if err := m.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetUserWithBilling retrieves user and their billing information
func (m *BillingModel) GetUserWithBilling(userID uint) (*UserBillingInfo, error) {
	var info UserBillingInfo

	user, err := m.userModel.FindByID(userID)
	if err != nil {
		return nil, err
	}
	info.User = user

	billingProfile, err := m.GetUserBillingProfile(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	info.BillingProfile = billingProfile

	return &info, nil
}

// UpdateSubscriptionStatus updates subscription and billing status
func (m *BillingModel) UpdateSubscriptionStatus(userID uint, status string) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		if err := user.UpdateSubscription(tx, status); err != nil {
			return err
		}

		now := time.Now()
		updates := map[string]interface{}{
			"billing_status":    BillingStatusActive,
			"last_billing_date": now,
		}

		var nextBillingDate time.Time
		var profile BillingProfile
		if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			return err
		}

		if profile.BillingCycle == BillingCycleMonthly {
			nextBillingDate = now.AddDate(0, 1, 0)
		} else {
			nextBillingDate = now.AddDate(1, 0, 0)
		}
		updates["next_billing_date"] = nextBillingDate

		return tx.Model(&profile).Updates(updates).Error
	})
}

// CancelSubscription cancels user's subscription
func (m *BillingModel) CancelSubscription(userID uint) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		if err := user.UpdateSubscription(tx, SubscriptionStatusCancelled); err != nil {
			return err
		}

		return tx.Model(&BillingProfile{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"billing_status":    BillingStatusCancelled,
				"next_billing_date": nil,
			}).Error
	})
}

// GetSubscriptionStats gets billing statistics
func (m *BillingModel) GetSubscriptionStats() (map[string]interface{}, error) {
	var stats map[string]interface{}

	err := m.db.Model(&BillingProfile{}).
		Select(`
            COUNT(*) as total_subscriptions,
            COUNT(CASE WHEN billing_status = 'active' THEN 1 END) as active_subscriptions,
            COUNT(CASE WHEN billing_cycle = 'monthly' THEN 1 END) as monthly_subscriptions,
            COUNT(CASE WHEN billing_cycle = 'yearly' THEN 1 END) as yearly_subscriptions
        `).
		Scan(&stats).Error

	return stats, err
}

// GetExpiringSubscriptions gets subscriptions expiring soon
func (m *BillingModel) GetExpiringSubscriptions(days int) ([]BillingProfile, error) {
	var profiles []BillingProfile
	expiryDate := time.Now().AddDate(0, 0, days)

	err := m.db.Where("billing_status = ? AND next_billing_date <= ?",
		BillingStatusActive, expiryDate).
		Find(&profiles).Error

	return profiles, err
}
