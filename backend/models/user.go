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
	Password           string     `json:"-" gorm:"not null"`
	Role               string     `json:"role" gorm:"type:enum('end_user','premium_user','sys_admin','super_admin');default:'end_user'"`
	ReadAccess         bool       `json:"read_access" gorm:"default:true"`
	WriteAccess        bool       `json:"write_access" gorm:"default:true"`
	TwoFactorEnabled   bool       `json:"two_factor_enabled" gorm:"default:false"`
	TwoFactorSecret    string     `json:"-" gorm:"column:two_factor_secret"`
	StorageQuota       int64      `json:"storage_quota" gorm:"default:5368709120"`
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

// View Sys admin account
func (m *UserModel) GetSysAdmins(creator *User) ([]*User, error) {
	if !creator.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only super admins can view system administrators")
	}

	var sysAdmins []*User
	err := m.db.Where("role = ?", RoleSysAdmin).Find(&sysAdmins).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching system administrators: %v", err)
	}

	return sysAdmins, nil
}

// Delete Sysadmin account
func (m *UserModel) DeleteSysAdmin(superAdmin *User, sysAdminID uint) error {
	if !superAdmin.IsSuperAdmin() {
		return errors.New("unauthorized: only super admins can delete system administrators")
	}

	// Get the system administrator to be deleted
	var sysAdmin User
	if err := m.db.First(&sysAdmin, sysAdminID).Error; err != nil {
		return errors.New("system administrator not found")
	}

	// Verify that the target user is actually a system administrator
	if !sysAdmin.IsSysAdmin() {
		return errors.New("specified user is not a system administrator")
	}

	// Prevent super admin from deleting themselves
	if sysAdmin.ID == superAdmin.ID {
		return errors.New("cannot delete your own account")
	}

	// Perform a soft delete (deactivate the account)
	sysAdmin.IsActive = false
	if err := m.db.Save(&sysAdmin).Error; err != nil {
		return fmt.Errorf("failed to deactivate system administrator: %v", err)
	}

	return nil
}

// UpdateUserAccess updates a user's access permissions
func (m *UserModel) UpdateUserAccess(sysAdmin *User, userID uint, readAccess, writeAccess bool) error {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return errors.New("unauthorized: only administrators can update user access")
	}

	var user User
	if err := m.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Prevent modification of admin accounts
	if user.IsSysAdmin() || user.IsSuperAdmin() {
		return errors.New("cannot modify administrator access permissions")
	}

	// Update access permissions
	if err := m.db.Model(&user).Updates(map[string]interface{}{
		"read_access":  readAccess,
		"write_access": writeAccess,
	}).Error; err != nil {
		return fmt.Errorf("failed to update user access: %v", err)
	}

	return nil
}

// DeleteUser performs a soft delete of a user account
func (m *UserModel) DeleteUser(sysAdmin *User, userID uint) error {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return errors.New("unauthorized: only administrators can delete user accounts")
	}

	var user User
	if err := m.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Prevent deletion of admin accounts
	if user.IsSysAdmin() || user.IsSuperAdmin() {
		return errors.New("cannot delete administrator accounts")
	}

	if err := user.DeactivateAccount(m.db); err != nil {
		return fmt.Errorf("failed to delete user account: %v", err)
	}

	return nil
}

// GetAllUsers retrieves all user accounts excluding administrators
func (m *UserModel) GetAllUsers(sysAdmin *User) ([]*User, error) {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only administrators can view all users")
	}

	var users []*User
	err := m.db.Where("role NOT IN ?", []string{RoleSysAdmin, RoleSuperAdmin}).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %v", err)
	}

	return users, nil
}

// GetStorageUsage retrieves storage usage statistics
func (m *UserModel) GetStorageUsage(sysAdmin *User) (map[string]interface{}, error) {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only administrators can view storage usage")
	}

	var totalUsed int64
	var totalQuota int64
	var userCount int64
	var activeUsers int64

	err := m.db.Model(&User{}).
		Where("role NOT IN ?", []string{RoleSysAdmin, RoleSuperAdmin}).
		Select("COUNT(*) as user_count, "+
			"SUM(storage_used) as total_used, "+
			"SUM(storage_quota) as total_quota, "+
			"SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END) as active_users").
		Row().Scan(&userCount, &totalUsed, &totalQuota, &activeUsers)

	if err != nil {
		return nil, fmt.Errorf("error fetching storage statistics: %v", err)
	}

	return map[string]interface{}{
		"total_users":   userCount,
		"active_users":  activeUsers,
		"storage_used":  totalUsed,
		"storage_quota": totalQuota,
		"usage_percent": float64(totalUsed) / float64(totalQuota) * 100,
	}, nil
}

// GetSubscriptionDetails retrieves subscription and billing information
func (m *UserModel) GetSubscriptionDetails(sysAdmin *User) ([]map[string]interface{}, error) {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only administrators can view subscription details")
	}

	var subscriptions []map[string]interface{}
	err := m.db.Model(&User{}).
		Select("subscription_status, COUNT(*) as count, "+
			"SUM(storage_quota) as total_quota").
		Where("role NOT IN ?", []string{RoleSysAdmin, RoleSuperAdmin}).
		Group("subscription_status").
		Scan(&subscriptions).Error

	if err != nil {
		return nil, fmt.Errorf("error fetching subscription details: %v", err)
	}

	return subscriptions, nil
}

// GetDeletedUsers retrieves all deleted user accounts
func (m *UserModel) GetDeletedUsers(sysAdmin *User) ([]*User, error) {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only administrators can view deleted users")
	}

	var users []*User
	err := m.db.Where("is_active = ?", false).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching deleted users: %v", err)
	}

	return users, nil
}

// RestoreUser restores a deleted user account
func (m *UserModel) RestoreUser(sysAdmin *User, userID uint) error {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return errors.New("unauthorized: only administrators can restore user accounts")
	}

	var user User
	if err := m.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if user.IsActive {
		return errors.New("user account is already active")
	}

	// Prevent restoration of admin accounts by non-super-admins
	if (user.IsSysAdmin() || user.IsSuperAdmin()) && !sysAdmin.IsSuperAdmin() {
		return errors.New("cannot restore administrator accounts")
	}

	return user.ReactivateAccount(m.db)
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
