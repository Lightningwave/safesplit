package models

import (
	"errors"
	"fmt"
	"safesplit/services"
	"safesplit/utils"
	"strconv"
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

type StorageStats struct {
	TotalUsers  int64                `json:"total_users"`
	ActiveUsers int64                `json:"active_users"`
	StorageUsed int64                `json:"storage_used"`
	Users       []UserStorageDetails `json:"users"`
}

type UserStorageDetails struct {
	ID                 string `json:"id"`
	Username           string `json:"username"`
	SubscriptionStatus string `json:"subscription_status"`
	StorageUsed        int64  `json:"storage_used"`
	StorageTotal       int64  `json:"storage_total"`
}

// DefaultStorageQuota represents 5GB in bytes for free users
const DefaultStorageQuota = int64(5 * 1024 * 1024 * 1024)

// PremiumStorageQuota represents 50GB in bytes for premium users
const PremiumStorageQuota = int64(50 * 1024 * 1024 * 1024)

type User struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	Username            string     `json:"username" gorm:"unique;not null"`
	Email               string     `json:"email" gorm:"unique;not null"`
	Password            string     `json:"-" gorm:"not null"`
	MasterKeySalt       []byte     `json:"-" gorm:"type:binary(32);not null"`
	MasterKeyNonce      []byte     `json:"-" gorm:"type:binary(16);not null"`
	EncryptedMasterKey  []byte     `json:"-" gorm:"type:binary(64);not null"`
	MasterKeyVersion    int        `json:"-" gorm:"not null;default:1"`
	KeyLastRotated      *time.Time `json:"-"`
	Role                string     `json:"role" gorm:"type:enum('end_user','premium_user','sys_admin','super_admin');default:'end_user'"`
	ReadAccess          bool       `json:"read_access" gorm:"default:true"`
	WriteAccess         bool       `json:"write_access" gorm:"default:true"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled" gorm:"default:false"`
	TwoFactorSecret     string     `json:"-" gorm:"column:two_factor_secret"`
	StorageQuota        int64      `json:"storage_quota" gorm:"default:5368709120"` // 5GB default
	StorageUsed         int64      `json:"storage_used" gorm:"default:0"`
	SubscriptionStatus  string     `json:"subscription_status" gorm:"type:enum('free','premium','cancelled');default:'free'"`
	IsActive            bool       `json:"is_active" gorm:"default:true"`
	LastLogin           *time.Time `json:"last_login"`
	LastPasswordChange  time.Time  `json:"last_password_change" gorm:"autoCreateTime"`
	FailedLoginAttempts int        `json:"failed_login_attempts" gorm:"default:0"`
	AccountLockedUntil  *time.Time `json:"account_locked_until"`
	ForcePasswordChange bool       `json:"force_password_change" gorm:"default:false"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserModel struct {
	db               *gorm.DB
	twoFactorService *services.TwoFactorAuthService
}

func NewUserModel(db *gorm.DB, twoFactorService *services.TwoFactorAuthService) *UserModel {
	return &UserModel{
		db:               db,
		twoFactorService: twoFactorService,
	}
}

// BeforeCreate hook to set up user security fields
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.Password = string(hashedPassword)

	// Generate master key
	masterKey, err := services.GenerateMasterKey()
	if err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	// Generate salt
	salt, err := utils.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	u.MasterKeySalt = salt

	// Generate nonce
	nonce, err := utils.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	u.MasterKeyNonce = nonce

	// Derive key encryption key from password using salt
	kek, err := services.DeriveKeyEncryptionKey(u.Password, salt)
	if err != nil {
		return fmt.Errorf("failed to derive key encryption key: %w", err)
	}

	// Encrypt master key
	encryptedKey, err := services.EncryptMasterKey(masterKey, kek, nonce)
	if err != nil {
		return fmt.Errorf("failed to encrypt master key: %w", err)
	}

	// Set encrypted master key and version
	u.EncryptedMasterKey = encryptedKey
	u.MasterKeyVersion = 1

	return nil
}

// Create creates a new user with master key generation
func (m *UserModel) Create(user *User) (*User, error) {
	var createdUser *User
	err := m.db.Transaction(func(tx *gorm.DB) error {
		// Generate salt and nonce (BeforeCreate hook will handle this)
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Generate and encrypt master key using password
		masterKey, err := services.GenerateMasterKey()
		if err != nil {
			return fmt.Errorf("failed to generate master key: %w", err)
		}

		// Derive key encryption key from password using salt
		kek, err := services.DeriveKeyEncryptionKey(user.Password, user.MasterKeySalt)
		if err != nil {
			return fmt.Errorf("failed to derive key encryption key: %w", err)
		}

		// Encrypt master key
		encryptedKey, err := services.EncryptMasterKey(masterKey, kek, user.MasterKeyNonce)
		if err != nil {
			return fmt.Errorf("failed to encrypt master key: %w", err)
		}

		// Update user with encrypted master key
		user.EncryptedMasterKey = encryptedKey
		if err := tx.Save(user).Error; err != nil {
			return fmt.Errorf("failed to save encrypted master key: %w", err)
		}

		createdUser = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// Authenticate checks the provided email and password
func (m *UserModel) Authenticate(email, password string) (*User, error) {
	var user User
	if err := m.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Prevent super admin login through regular endpoint
	if user.Role == RoleSuperAdmin {
		return nil, errors.New("please use super admin login portal")
	}

	// Check if account is locked
	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("account locked until %v", user.AccountLockedUntil)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if err := m.handleFailedLogin(&user); err != nil {
			return nil, err
		}
		return nil, errors.New("invalid credentials")
	}

	// Reset failed attempts and update login time
	return m.handleSuccessfulLogin(&user)
}

// handleSuccessfulLogin updates user state after successful login
func (m *UserModel) handleSuccessfulLogin(user *User) (*User, error) {
	user.FailedLoginAttempts = 0
	user.AccountLockedUntil = nil
	now := time.Now()
	user.LastLogin = &now

	if err := m.db.Save(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// handleFailedLogin manages failed login attempts and account lockout
func (m *UserModel) handleFailedLogin(user *User) error {
	user.FailedLoginAttempts++

	if user.FailedLoginAttempts >= 5 {
		lockTime := time.Now().Add(30 * time.Minute)
		user.AccountLockedUntil = &lockTime
	}

	return m.db.Save(user).Error
}

// UpdateMasterKey updates the user's master key material
func (u *User) UpdateMasterKey(db *gorm.DB, newEncryptedKey []byte) error {
	if len(newEncryptedKey) != 64 {
		return errors.New("invalid master key length")
	}

	// Generate new nonce
	nonce, err := utils.GenerateNonce()
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"encrypted_master_key": newEncryptedKey,
		"master_key_nonce":     nonce,
		"master_key_version":   u.MasterKeyVersion + 1,
		"key_last_rotated":     now,
	}

	if err := db.Model(u).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update master key: %w", err)
	}

	// Update local struct
	u.EncryptedMasterKey = newEncryptedKey
	u.MasterKeyNonce = nonce
	u.MasterKeyVersion++
	u.KeyLastRotated = &now

	return nil
}

// RotateMasterKey performs a key rotation operation
func (m *UserModel) RotateMasterKey(userID uint, newEncryptedKey []byte, rotationType RotationType, keyRotationModel *KeyRotationModel) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		// Get user and verify existence
		var user User
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// Verify rotation type is valid
		switch rotationType {
		case RotationTypeAutomatic, RotationTypeManual, RotationTypeForced, RotationTypePassword:
			// Valid rotation type
		default:
			return fmt.Errorf("invalid rotation type: %s", rotationType)
		}

		// Store old version for logging
		oldVersion := user.MasterKeyVersion

		// Update the master key
		if err := user.UpdateMasterKey(tx, newEncryptedKey); err != nil {
			return fmt.Errorf("failed to update master key: %w", err)
		}

		// Log the rotation using the KeyRotationModel
		if err := keyRotationModel.LogRotation(userID, oldVersion, user.MasterKeyVersion, rotationType); err != nil {
			return fmt.Errorf("failed to log rotation: %w", err)
		}

		return nil
	})
}

func (m *UserModel) AuthenticateSuperAdmin(email, password string) (*User, error) {
	var user User
	if err := m.db.Where("email = ? AND is_active = ? AND role = ?",
		email, true, RoleSuperAdmin).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if account is locked
	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("account locked until %v", user.AccountLockedUntil)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if err := m.handleFailedLogin(&user); err != nil {
			return nil, err
		}
		return nil, errors.New("invalid credentials")
	}

	return m.handleSuccessfulLogin(&user)
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

// ResetPassword updates the user's password and optionally rotates master key
func (m *UserModel) ResetPassword(userID uint, currentPassword, newPassword string, newEncryptedMasterKey []byte, passwordHistoryModel *PasswordHistoryModel) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.First(&user, userID).Error; err != nil {
			return errors.New("user not found")
		}

		// Verify current password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
			return errors.New("current password is incorrect")
		}

		// Generate new password hash
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// Store the old password in history
		if err := passwordHistoryModel.AddEntry(user.ID, user.Password); err != nil {
			return err
		}

		// Update password and related fields
		now := time.Now()
		updates := map[string]interface{}{
			"password":              string(hashedPassword),
			"last_password_change":  now,
			"force_password_change": false,
		}

		// Update master key if provided
		if newEncryptedMasterKey != nil && len(newEncryptedMasterKey) > 0 {
			nonce, err := utils.GenerateNonce()
			if err != nil {
				return err
			}

			updates["encrypted_master_key"] = newEncryptedMasterKey
			updates["master_key_nonce"] = nonce
			updates["master_key_version"] = user.MasterKeyVersion + 1
			updates["key_last_rotated"] = now

			// Log key rotation
			rotation := KeyRotationHistory{
				UserID:        userID,
				OldKeyVersion: user.MasterKeyVersion,
				NewKeyVersion: user.MasterKeyVersion + 1,
				RotationType:  "password_change",
			}

			if err := tx.Create(&rotation).Error; err != nil {
				return fmt.Errorf("failed to log key rotation: %w", err)
			}
		}

		return tx.Model(&user).Updates(updates).Error
	})
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

	if err := m.DeactivateAccount(userID); err != nil {
		return fmt.Errorf("failed to delete user account: %w", err)
	}

	return nil
}

// GetAllUsers retrieves all user accounts excluding administrators
func (m *UserModel) GetAllUsers(sysAdmin *User) ([]*User, error) {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only administrators can view all users")
	}

	var users []*User
	err := m.db.Where("role NOT IN ?", []string{RoleSysAdmin, RoleSuperAdmin}).
		Where("is_active = ?", true).
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %v", err)
	}

	return users, nil
}

// GetStorageUsage retrieves storage usage statistics
func (m *UserModel) GetStorageStats() (*StorageStats, error) {
	var stats StorageStats
	var users []User

	// Get all users excluding admins
	if err := m.db.Where("role NOT IN ?", []string{RoleSysAdmin, RoleSuperAdmin}).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("error fetching users: %v", err)
	}

	// Calculate totals and prepare user details
	userDetails := make([]UserStorageDetails, 0, len(users))
	for _, user := range users {
		stats.TotalUsers++
		if user.IsActive {
			stats.ActiveUsers++
			stats.StorageUsed += user.StorageUsed

			userDetails = append(userDetails, UserStorageDetails{
				ID:                 strconv.FormatUint(uint64(user.ID), 10),
				Username:           user.Username,
				SubscriptionStatus: user.SubscriptionStatus,
				StorageUsed:        user.StorageUsed,
				StorageTotal:       user.StorageQuota,
			})
		}
	}

	stats.Users = userDetails
	return &stats, nil
}

// DeactivateAccount deactivates the user account
func (m *UserModel) DeactivateAccount(userID uint) error {
	result := m.db.Model(&User{}).
		Where("id = ?", userID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found or already deactivated")
	}

	return nil
}

// ReactivateAccount reactivates the user account
func (m *UserModel) ReactivateAccount(userID uint) error {
	result := m.db.Model(&User{}).
		Where("id = ?", userID).
		Update("is_active", true)

	if result.Error != nil {
		return fmt.Errorf("failed to reactivate account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found or already active")
	}

	return nil
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

	return m.ReactivateAccount(userID)
}

// UpdateUserAccount updates a user's account information and privileges
func (m *UserModel) UpdateUserAccount(sysAdmin *User, userID uint, updates *User) error {
	if !sysAdmin.IsSysAdmin() && !sysAdmin.IsSuperAdmin() {
		return errors.New("unauthorized: only administrators can update user accounts")
	}

	var user User
	if err := m.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Prevent modification of admin accounts
	if user.IsSysAdmin() || user.IsSuperAdmin() {
		return errors.New("cannot modify administrator accounts")
	}

	// Update basic information
	user.Username = updates.Username
	user.Email = updates.Email
	user.ReadAccess = updates.ReadAccess
	user.WriteAccess = updates.WriteAccess

	// Handle subscription status change if needed
	if user.SubscriptionStatus != updates.SubscriptionStatus {
		if err := user.UpdateSubscription(m.db, updates.SubscriptionStatus); err != nil {
			return fmt.Errorf("failed to update subscription: %v", err)
		}
	}

	if err := m.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user account: %v", err)
	}

	return nil
}

func (m *UserModel) UpdateUserStorage(userID uint, size int64) error {
	var user User
	if err := m.db.First(&user, userID).Error; err != nil {
		return err
	}

	return m.db.Model(&user).Update("storage_used", user.StorageUsed+size).Error
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

// InitiateEmailTwoFactor starts the email 2FA process
func (m *UserModel) InitiateEmailTwoFactor(userID uint) error {
	user, err := m.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return errors.New("account is not active")
	}

	return m.twoFactorService.SendTwoFactorToken(userID, user.Email)
}

// VerifyEmailTwoFactor verifies the email 2FA token
func (m *UserModel) VerifyEmailTwoFactor(userID uint, token string) error {
	if err := m.twoFactorService.VerifyToken(userID, token); err != nil {
		return err
	}

	return m.db.Model(&User{}).Where("id = ?", userID).Update("two_factor_enabled", true).Error
}

// EnableEmailTwoFactor enables email-based 2FA
func (m *UserModel) EnableEmailTwoFactor(userID uint) error {
	return m.db.Model(&User{}).
		Where("id = ?", userID).
		Update("two_factor_enabled", true).Error
}

// DisableEmailTwoFactor disables email-based 2FA
func (m *UserModel) DisableEmailTwoFactor(userID uint) error {
	return m.db.Model(&User{}).
		Where("id = ?", userID).
		Update("two_factor_enabled", false).Error
}
