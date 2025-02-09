package models

import (
	"errors"
	"time"
	"fmt"

	"gorm.io/gorm"
)

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
var ErrStorageExceedsQuota = errors.New("cannot downgrade: storage usage exceeds free tier quota")


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
        var existingProfile BillingProfile
        if err := tx.Where("user_id = ?", profile.UserID).First(&existingProfile).Error; err == nil {
            return errors.New("user already has a billing profile")
        }

        if profile.CustomerID == "" {
            profile.CustomerID = fmt.Sprintf("CUST_%d_%s", profile.UserID, time.Now().Format("20060102"))
        }

        if err := tx.Create(profile).Error; err != nil {
            return fmt.Errorf("failed to create billing profile: %v", err)
        }

        return nil
    })
}

func (m *BillingModel) UpdateBillingProfile(profile *BillingProfile) error {
    return m.db.Transaction(func(tx *gorm.DB) error {
        updates := map[string]interface{}{
            "billing_name": profile.BillingName,
            "billing_email": profile.BillingEmail,
            "billing_address": profile.BillingAddress,
            "country_code": profile.CountryCode,
            "billing_cycle": profile.BillingCycle,
            "billing_status": profile.BillingStatus,
            "default_payment_method": profile.DefaultPaymentMethod,
            "currency": profile.Currency,
        }
        
        result := tx.Model(&BillingProfile{}).
            Where("user_id = ?", profile.UserID).
            Updates(updates)
            
        if result.Error != nil {
            return fmt.Errorf("database error: %v", result.Error)
        }
        
        if result.RowsAffected == 0 {
            return fmt.Errorf("no profile found for user_id: %d", profile.UserID)
        }
        
        return nil
    })
}
func (m *BillingModel) GetUserBillingProfile(userID uint) (*BillingProfile, error) {
	var profile BillingProfile
	if err := m.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

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

func (m *BillingModel) CancelSubscription(userID uint) error {
    return m.db.Transaction(func(tx *gorm.DB) error {
        var user User
        if err := tx.First(&user, userID).Error; err != nil {
            return err
        }

        if user.StorageUsed > DefaultStorageQuota {
            return ErrStorageExceedsQuota
        }

        var profile BillingProfile
        if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
            return err
        }

        return tx.Model(&profile).
            Updates(map[string]interface{}{
                "billing_status": BillingStatusCancelled,
            }).Error
    })
}

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

func (m *BillingModel) GetExpiringSubscriptions(days int) ([]BillingProfile, error) {
	var profiles []BillingProfile
	expiryDate := time.Now().AddDate(0, 0, days)

	err := m.db.Where("billing_status = ? AND next_billing_date <= ?",
		BillingStatusActive, expiryDate).
		Find(&profiles).Error

	return profiles, err
}

