package models

import (
	"errors"
	"fmt"
	"log"
	"safesplit/services"
	"safesplit/utils"
	"strconv"
	"strings"
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.Password = string(hashedPassword)

	masterKey, err := services.GenerateMasterKey()
	if err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	salt, err := utils.GenerateSalt()
	if err != nil {
		return err
	}
	u.MasterKeySalt = salt

	nonce, err := utils.GenerateNonce()
	if err != nil {
		return err
	}
	u.MasterKeyNonce = nonce

	// Use hashed password for KEK derivation
	kek, err := services.DeriveKeyEncryptionKey(string(hashedPassword), salt)
	if err != nil {
		return fmt.Errorf("failed to derive key encryption key: %w", err)
	}

	encryptedKey, err := services.EncryptMasterKey(masterKey, kek, nonce)
	if err != nil {
		return fmt.Errorf("failed to encrypt master key: %w", err)
	}
	u.EncryptedMasterKey = encryptedKey
	u.MasterKeyVersion = 1

	return nil
}

// Create creates a new user with master key generation
func (m *UserModel) Create(user *User) (*User, error) {
	var createdUser *User
	err := m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		createdUser = user
		return nil
	})
	return createdUser, err
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
		remainingTime := user.AccountLockedUntil.Sub(time.Now())
		return nil, fmt.Errorf("account locked for %d minutes due to too many failed attempts", int(remainingTime.Minutes()))
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		lockErr := m.handleFailedLogin(&user)
		if lockErr != nil && strings.Contains(lockErr.Error(), "account locked") {
			return nil, lockErr
		}
		remainingAttempts := 5 - user.FailedLoginAttempts
		if remainingAttempts > 0 {
			return nil, fmt.Errorf("invalid credentials - %d attempts remaining", remainingAttempts)
		}
		return nil, errors.New("invalid credentials")
	}

	updatedUser, err := m.handleSuccessfulLogin(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to update login status: %w", err)
	}

	if updatedUser.Role == RoleSysAdmin || updatedUser.TwoFactorEnabled {
		return updatedUser, nil
	}

	return updatedUser, nil
}

// handleSuccessfulLogin updates user state after successful login
func (m *UserModel) handleSuccessfulLogin(user *User) (*User, error) {
	user.FailedLoginAttempts = 0
	user.AccountLockedUntil = nil
	now := time.Now()
	user.LastLogin = &now

	if err := m.db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update login state: %w", err)
	}

	return user, nil
}

// handleFailedLogin manages failed login attempts and account lockout
func (m *UserModel) handleFailedLogin(user *User) error {
	user.FailedLoginAttempts++

	var lockTime *time.Time
	if user.FailedLoginAttempts >= 5 {
		t := time.Now().Add(30 * time.Minute)
		lockTime = &t
		user.AccountLockedUntil = lockTime
	}

	if err := m.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update login attempts: %w", err)
	}

	if lockTime != nil {
		return fmt.Errorf("account locked until %v due to too many failed attempts", lockTime)
	}

	return nil
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

func (m *UserModel) AuthenticateSuperAdmin(email, password string) (*User, error) {
	var user User
	if err := m.db.Where("email = ? AND is_active = ? AND role = ?",
		email, true, RoleSuperAdmin).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		remainingTime := user.AccountLockedUntil.Sub(time.Now())
		return nil, fmt.Errorf("account locked for %d minutes due to too many failed attempts", int(remainingTime.Minutes()))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		lockErr := m.handleFailedLogin(&user)
		if lockErr != nil && strings.Contains(lockErr.Error(), "account locked") {
			return nil, lockErr
		}
		remainingAttempts := 5 - user.FailedLoginAttempts
		if remainingAttempts > 0 {
			return nil, fmt.Errorf("invalid credentials - %d attempts remaining", remainingAttempts)
		}
		return nil, errors.New("invalid credentials")
	}

	updatedUser, err := m.handleSuccessfulLogin(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to update login status: %w", err)
	}

	return updatedUser, nil
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

// Create Sys admin account
func (m *UserModel) CreateSysAdmin(creator *User, newAdmin *User) (*User, error) {
	if !creator.IsSuperAdmin() {
		return nil, errors.New("unauthorized: only super admins can create system administrators")
	}

	err := m.db.Transaction(func(tx *gorm.DB) error {
		newAdmin.Role = RoleSysAdmin

		if err := tx.Create(newAdmin).Error; err != nil {
			return fmt.Errorf("failed to create system administrator: %v", err)
		}

		if err := tx.Model(newAdmin).Update("two_factor_enabled", true).Error; err != nil {
			return fmt.Errorf("failed to enable 2FA: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var createdAdmin User
	if err := m.db.First(&createdAdmin, newAdmin.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load created admin: %v", err)
	}

	return &createdAdmin, nil
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

func (m *UserModel) ResetPasswordWithFragments(
	userID uint,
	currentPassword string,
	newPassword string,
	passwordHistoryModel *PasswordHistoryModel,
	keyFragmentModel *KeyFragmentModel,
	fileModel *FileModel,
) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// Verify current password first
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
			return fmt.Errorf("current password is incorrect")
		}

		// Hash new password FIRST to check for reuse
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash new password: %w", err)
		}

		// Check if the new password matches any of the recent passwords
		recentPasswords, err := passwordHistoryModel.GetRecentPasswords(userID, 5)
		if err != nil {
			return fmt.Errorf("failed to check password history: %w", err)
		}

		// Compare new password hash with recent password hashes
		for _, oldHash := range recentPasswords {
			if err := bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(newPassword)); err == nil {
				// If CompareHashAndPassword returns nil, the password matches
				return errors.New("Cannot reuse any of your last 5 passwords")
			}
		}

		// Store current password in history BEFORE updating to new one
		if err := passwordHistoryModel.AddEntry(user.ID, user.Password); err != nil {
			return fmt.Errorf("failed to store password history: %w", err)
		}

		// Get original encrypted master key
		originalEncryptedKey := user.EncryptedMasterKey
		log.Printf("Current encrypted master key: %x", originalEncryptedKey)

		// Derive current KEK
		currentKEK, err := services.DeriveKeyEncryptionKey(user.Password, user.MasterKeySalt)
		if err != nil {
			return fmt.Errorf("failed to derive current KEK: %w", err)
		}

		// First decrypt master key with current KEK
		decryptedMasterKey, err := services.DecryptMasterKey(originalEncryptedKey, currentKEK, user.MasterKeyNonce)
		if err != nil {
			return fmt.Errorf("failed to decrypt master key: %w", err)
		}

		// Derive new KEK using already hashed password
		newKEK, err := services.DeriveKeyEncryptionKey(string(hashedNewPassword), user.MasterKeySalt)
		if err != nil {
			return fmt.Errorf("failed to derive new KEK: %w", err)
		}

		// Generate new nonce for master key
		masterKeyNonce, err := utils.GenerateNonce()
		if err != nil {
			return fmt.Errorf("failed to generate nonce: %w", err)
		}

		// Re-encrypt master key with new KEK
		newEncryptedMasterKey, err := services.EncryptMasterKey(decryptedMasterKey, newKEK, masterKeyNonce)
		if err != nil {
			return fmt.Errorf("failed to re-encrypt master key: %w", err)
		}

		log.Printf("New encrypted master key: %x", newEncryptedMasterKey)

		// Store old password in history BEFORE updating
		if err := passwordHistoryModel.AddEntry(user.ID, user.Password); err != nil {
			return fmt.Errorf("failed to store password history: %w", err)
		}

		// Process fragments
		files, err := fileModel.ListAllUserFiles(userID)
		if err != nil {
			return fmt.Errorf("failed to get user files: %w", err)
		}

		// Use decrypted master key for fragments
		userMasterKey := decryptedMasterKey[:32]

		for _, file := range files {
			fragments, err := keyFragmentModel.GetUserFragmentsForFile(file.ID)
			if err != nil {
				if err.Error() == "record not found" {
					log.Printf("Warning: skipping missing fragments for file %d", file.ID)
					continue
				}
				return fmt.Errorf("failed to get key fragments for file %d: %w", file.ID, err)
			}

			for _, fragment := range fragments {
				log.Printf("Processing fragment %d for file %d", fragment.FragmentIndex, file.ID)

				// Decrypt fragment with current decrypted master key
				decryptedFragment, err := services.DecryptMasterKey(
					fragment.Data,
					userMasterKey,
					fragment.EncryptionNonce,
				)
				if err != nil {
					log.Printf("Warning: skipping unreadable fragment %d for file %d: %v",
						fragment.FragmentIndex, file.ID, err)
					continue
				}

				// Generate new nonce for fragment
				newFragmentNonce, err := utils.GenerateNonce()
				if err != nil {
					return fmt.Errorf("failed to generate nonce for fragment: %w", err)
				}

				// Re-encrypt with same decrypted master key
				newEncryptedFragment, err := services.EncryptMasterKey(
					decryptedFragment,
					userMasterKey,
					newFragmentNonce,
				)
				if err != nil {
					return fmt.Errorf("failed to re-encrypt fragment: %w", err)
				}

				// Store re-encrypted fragment
				if err := keyFragmentModel.storage.StoreFragment(
					fragment.NodeIndex,
					fragment.FragmentPath,
					newEncryptedFragment,
				); err != nil {
					return fmt.Errorf("failed to store fragment: %w", err)
				}

				// Update fragment metadata
				newVersion := user.MasterKeyVersion + 1
				if err := tx.Model(&fragment.KeyFragment).Updates(map[string]interface{}{
					"encryption_nonce":   newFragmentNonce,
					"master_key_version": newVersion,
				}).Error; err != nil {
					return fmt.Errorf("failed to update fragment metadata: %w", err)
				}
			}
		}

		now := time.Now()
		updates := map[string]interface{}{
			"password":              string(hashedNewPassword),
			"encrypted_master_key":  newEncryptedMasterKey,
			"master_key_nonce":      masterKeyNonce,
			"master_key_version":    user.MasterKeyVersion + 1,
			"key_last_rotated":      now,
			"last_password_change":  now,
			"force_password_change": false,
		}

		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}
func (m *UserModel) updateKeyFragments(
	tx *gorm.DB,
	userID uint,
	oldMasterKey []byte,
	newMasterKey []byte,
	keyFragmentModel *KeyFragmentModel,
	fileModel *FileModel,
) error {
	files, err := fileModel.ListAllUserFiles(userID)
	if err != nil {
		return fmt.Errorf("failed to get user files: %w", err)
	}

	// Get the decrypted master key for the user
	decryptedMasterKey := newMasterKey
	for _, file := range files {
		fragments, err := keyFragmentModel.GetUserFragmentsForFile(file.ID)
		if err != nil {
			return fmt.Errorf("failed to get key fragments for file %d: %w", file.ID, err)
		}

		for _, fragment := range fragments {
			log.Printf("Processing fragment %d for file %d", fragment.FragmentIndex, file.ID)
			log.Printf("Fragment nonce length: %d", len(fragment.EncryptionNonce))
			log.Printf("Fragment data length: %d", len(fragment.Data))

			// Decrypt fragment using old master key
			decryptedFragment, err := services.DecryptMasterKey(
				fragment.Data,
				oldMasterKey,
				fragment.EncryptionNonce,
			)
			if err != nil {
				return fmt.Errorf("failed to decrypt fragment %d for file %d: %w",
					fragment.FragmentIndex, file.ID, err)
			}

			// Generate new nonce
			newNonce, err := utils.GenerateNonce()
			if err != nil {
				return fmt.Errorf("failed to generate nonce: %w", err)
			}

			log.Printf("Re-encrypting fragment with decrypted master key")

			// Re-encrypt fragment with new decrypted master key
			newEncryptedFragment, err := services.EncryptMasterKey(
				decryptedFragment,
				decryptedMasterKey,
				newNonce,
			)
			if err != nil {
				return fmt.Errorf("failed to re-encrypt fragment: %w", err)
			}

			// Store updated fragment
			if err := keyFragmentModel.storage.StoreFragment(
				fragment.NodeIndex,
				fragment.FragmentPath,
				newEncryptedFragment,
			); err != nil {
				return fmt.Errorf("failed to store fragment: %w", err)
			}

			// Update fragment metadata
			updates := map[string]interface{}{
				"encryption_nonce":   newNonce,
				"master_key_version": gorm.Expr("master_key_version + ?", 1),
			}

			if err := tx.Model(&fragment.KeyFragment).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update fragment metadata: %w", err)
			}

			log.Printf("Successfully updated fragment %d for file %d",
				fragment.FragmentIndex, file.ID)
		}
	}

	return nil
}