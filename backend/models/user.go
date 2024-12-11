package models

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Subscription status constants
const (
	SubscriptionStatusFree      = "free"
	SubscriptionStatusPremium   = "premium"
	SubscriptionStatusCancelled = "cancelled"
)

// Define user roles as constants
const (
	RoleEndUser     = "end_user"
	RolePremiumUser = "premium_user"
	RoleSysAdmin    = "sys_admin"
	RoleSuperAdmin  = "super_admin"
)

// DefaultStorageQuota represents 5GB in bytes for free users
const DefaultStorageQuota = int64(5 * 1024 * 1024 * 1024)

// PremiumStorageQuota represents 50GB in bytes for premium users
const PremiumStorageQuota = int64(50 * 1024 * 1024 * 1024)

type User struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	Username           string     `json:"username" gorm:"unique;not null"`
	Email              string     `json:"email" gorm:"unique;not null"`
	Password           string     `json:"-" gorm:"not null"` // "-" to exclude from JSON
	Role               string     `json:"role" gorm:"type:enum('end_user','premium_user','sys_admin','super_admin');default:'end_user'"`
	TwoFactorEnabled   bool       `json:"two_factor_enabled" gorm:"default:false"`
	TwoFactorSecret    string     `json:"-" gorm:"column:two_factor_secret"`       // "-" for security
	StorageQuota       int64      `json:"storage_quota" gorm:"default:5368709120"` // 5GB default
	StorageUsed        int64      `json:"storage_used" gorm:"default:0"`
	SubscriptionStatus string     `json:"subscription_status" gorm:"type:enum('free','premium','cancelled');default:'free'"`
	IsActive           bool       `json:"is_active" gorm:"default:true"`
	LastLogin          *time.Time `json:"last_login"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserModel struct {
	db *gorm.DB
}

func NewUserModel(db *gorm.DB) *UserModel {
	return &UserModel{db: db}
}

// BeforeCreate hook to hash password before saving
func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// Create creates a new user in the database
func (m *UserModel) Create(user *User) (*User, error) {
	if err := m.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// Authenticate checks the provided email and password
func (m *UserModel) Authenticate(email, password string) (*User, error) {
	var user User
	if err := m.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	if err := m.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByEmail retrieves a user by their email
func (m *UserModel) FindByEmail(email string) (*User, error) {
	var user User
	if err := m.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by their ID
func (m *UserModel) FindByID(id uint) (*User, error) {
	var user User
	if err := m.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateStorageUsed updates the user's storage usage
func (u *User) UpdateStorageUsed(db *gorm.DB, size int64) error {
	newUsage := u.StorageUsed + size
	if newUsage > u.StorageQuota {
		return errors.New("storage quota exceeded")
	}
	u.StorageUsed = newUsage
	return db.Save(u).Error
}

// Enable2FA enables two-factor authentication for the user
func (u *User) Enable2FA(db *gorm.DB, secret string) error {
	u.TwoFactorEnabled = true
	u.TwoFactorSecret = secret
	return db.Save(u).Error
}

// Disable2FA disables two-factor authentication for the user
func (u *User) Disable2FA(db *gorm.DB) error {
	u.TwoFactorEnabled = false
	u.TwoFactorSecret = ""
	return db.Save(u).Error
}

// UpdateSubscription updates the user's subscription status and storage quota
func (u *User) UpdateSubscription(db *gorm.DB, status string) error {
	u.SubscriptionStatus = status
	if status == SubscriptionStatusPremium {
		u.StorageQuota = PremiumStorageQuota
		u.Role = RolePremiumUser
	} else {
		if u.StorageUsed > DefaultStorageQuota {
			return errors.New("cannot downgrade: storage usage exceeds free tier quota")
		}
		u.StorageQuota = DefaultStorageQuota
		u.Role = RoleEndUser
	}
	return db.Save(u).Error
}

// DeactivateAccount deactivates the user account
func (u *User) DeactivateAccount(db *gorm.DB) error {
	u.IsActive = false
	return db.Save(u).Error
}

// ReactivateAccount reactivates the user account
func (u *User) ReactivateAccount(db *gorm.DB) error {
	u.IsActive = true
	return db.Save(u).Error
}

// ChangePassword updates the user's password
func (u *User) ChangePassword(db *gorm.DB, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return db.Save(u).Error
}

// Create Sys admin account
func (m *UserModel) CreateSysAdmin(creator *User, newAdmin *User) (*User, error) {
	if !creator.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only super admins can create system administrators")
	}

	// Ensure the new user is created as a sys_admin
	newAdmin.Role = RoleSysAdmin

	// Create the new admin user
	if err := m.db.Create(newAdmin).Error; err != nil {
		return nil, fmt.Errorf("failed to create system administrator: %v", err)
	}

	return newAdmin, nil
}

// Role check methods
func (u *User) IsEndUser() bool {
	return u.Role == RoleEndUser
}

func (u *User) IsPremiumUser() bool {
	return u.Role == RolePremiumUser
}

func (u *User) IsSysAdmin() bool {
	return u.Role == RoleSysAdmin
}

func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// HasAvailableStorage checks if the user has enough storage for the given size
func (u *User) HasAvailableStorage(size int64) bool {
	return u.StorageUsed+size <= u.StorageQuota
}
