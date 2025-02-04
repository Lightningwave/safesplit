package models

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type FileShare struct {
	ID                   uint       `json:"id" gorm:"primaryKey"`
	FileID               uint       `json:"file_id"`
	SharedBy             uint       `json:"shared_by"`
	ShareLink            string     `json:"share_link" gorm:"unique"`
	PasswordHash         string     `json:"-"`
	PasswordSalt         string     `json:"-"`
	EncryptedKeyFragment []byte     `json:"-" gorm:"type:mediumblob"`
	FragmentIndex        int        `json:"-" gorm:"not null"`
	ExpiresAt            *time.Time `json:"expires_at"`
	MaxDownloads         *int       `json:"max_downloads"`
	DownloadCount        int        `json:"download_count" gorm:"default:0"`
	IsActive             bool       `json:"is_active" gorm:"default:true"`
	CreatedAt            time.Time  `json:"created_at"`
	File                 File       `json:"file" gorm:"foreignKey:FileID"`
}

type FileShareModel struct {
	db *gorm.DB
}

func NewFileShareModel(db *gorm.DB) *FileShareModel {
	return &FileShareModel{db: db}
}

func generateShareLink() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (m *FileShareModel) CreateFileShareWithStatus(share *FileShare, password string) error {
	// Generate password salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	share.PasswordSalt = base64.StdEncoding.EncodeToString(salt)

	// Hash password with salt
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password+share.PasswordSalt),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	share.PasswordHash = string(hashedPassword)

	// Generate unique share link
	shareLink, err := generateShareLink()
	if err != nil {
		return fmt.Errorf("failed to generate share link: %w", err)
	}
	share.ShareLink = shareLink

	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Create share record within transaction
	if err := tx.Create(share).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create share record: %w", err)
	}

	// Update file's IsShared status
	if err := tx.Model(&File{}).Where("id = ?", share.FileID).Update("is_shared", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update file status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *FileShareModel) ValidateShareAccess(shareLink string, password string) (*FileShare, error) {
	var share FileShare
	if err := m.db.Where("share_link = ? AND is_active = ?", shareLink, true).
		Preload("File").First(&share).Error; err != nil {
		return nil, fmt.Errorf("share not found or inactive")
	}

	// Check expiration
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		share.IsActive = false
		m.db.Save(&share)
		return nil, fmt.Errorf("share has expired")
	}

	// Check download limit
	if share.MaxDownloads != nil && share.DownloadCount >= *share.MaxDownloads {
		share.IsActive = false
		m.db.Save(&share)
		return nil, fmt.Errorf("download limit exceeded")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(share.PasswordHash),
		[]byte(password+share.PasswordSalt),
	); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &share, nil
}

// CreateFileShare creates a basic file share with just password protection
func (m *FileShareModel) CreateFileShare(share *FileShare, password string) error {
	// Generate password salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	share.PasswordSalt = base64.StdEncoding.EncodeToString(salt)

	// Hash password with salt
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password+share.PasswordSalt),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	share.PasswordHash = string(hashedPassword)

	// Generate unique share link
	shareLink, err := generateShareLink()
	if err != nil {
		return fmt.Errorf("failed to generate share link: %w", err)
	}
	share.ShareLink = shareLink

	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Create share record within transaction
	if err := tx.Create(share).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create share record: %w", err)
	}

	// Update file's IsShared status
	if err := tx.Model(&File{}).Where("id = ?", share.FileID).Update("is_shared", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update file status: %w", err)
	}

	return tx.Commit().Error
}

// ValidateShare validates a share without checking expiry or download count
func (m *FileShareModel) ValidateShare(shareLink string, password string) (*FileShare, error) {
	var share FileShare
	if err := m.db.Where("share_link = ? AND is_active = ?", shareLink, true).
		Preload("File").First(&share).Error; err != nil {
		return nil, fmt.Errorf("share not found or inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(share.PasswordHash),
		[]byte(password+share.PasswordSalt),
	); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &share, nil
}

func (m *FileShareModel) IncrementDownloadCount(shareID uint) error {
	log.Printf("Starting IncrementDownloadCount for share ID %d", shareID)

	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to start transaction: %v", tx.Error)
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer tx.Rollback() // rollback if not committed

	log.Printf("Started transaction for share ID %d", shareID)

	result := tx.Model(&FileShare{}).
		Where("id = ?", shareID).
		Update("download_count", gorm.Expr("download_count + ?", 1))

	if result.Error != nil {
		log.Printf("Error during update: %v", result.Error)
		return fmt.Errorf("failed to increment download count: %w", result.Error)
	}

	log.Printf("Update query executed, affected rows: %d", result.RowsAffected)

	if result.RowsAffected == 0 {
		log.Printf("No rows affected for share ID %d", shareID)
		return fmt.Errorf("no share found with ID %d", shareID)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully committed download count increment for share ID %d", shareID)
	return nil
}
