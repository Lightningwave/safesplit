package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"safesplit/utils"
	"time"

	"gorm.io/gorm"
)

type ServerMasterKey struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	KeyID        string     `json:"key_id" gorm:"type:varchar(64);unique;not null"`
	EncryptedKey []byte     `json:"-" gorm:"type:binary(32);not null"`
	KeyNonce     []byte     `json:"-" gorm:"type:binary(16);not null"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ActivatedAt  *time.Time `json:"activated_at"`
	RetiredAt    *time.Time `json:"retired_at"`
}

type ServerMasterKeyModel struct {
	db *gorm.DB
}

func NewServerMasterKeyModel(db *gorm.DB) *ServerMasterKeyModel {
	return &ServerMasterKeyModel{db: db}
}

func generateKeyID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate key ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Initialize generates and stores the first server master key if none exists
func (m *ServerMasterKeyModel) Initialize() error {
	var count int64
	if err := m.db.Model(&ServerMasterKey{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing keys: %w", err)
	}

	if count > 0 {
		return nil
	}

	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	keyID, err := generateKeyID()
	if err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}

	nonce, err := utils.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	now := time.Now()
	serverKey := &ServerMasterKey{
		KeyID:        keyID,
		EncryptedKey: masterKey,
		KeyNonce:     nonce,
		IsActive:     true,
		ActivatedAt:  &now,
	}

	return m.db.Create(serverKey).Error
}

// GetServerKey retrieves and processes the server key for encryption
func (m *ServerMasterKeyModel) GetServerKey(keyID string) ([]byte, error) {
	var key ServerMasterKey
	if err := m.db.Where("key_id = ?", keyID).First(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to get server key: %w", err)
	}

	if len(key.EncryptedKey) != 32 {
		return nil, fmt.Errorf("invalid key length: got %d, expected 32 bytes", len(key.EncryptedKey))
	}

	return key.EncryptedKey, nil
}

// GetActive retrieves the current active server master key
func (m *ServerMasterKeyModel) GetActive() (*ServerMasterKey, error) {
	var key ServerMasterKey
	if err := m.db.Where("is_active = ? AND retired_at IS NULL", true).First(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to get active server key: %w", err)
	}
	return &key, nil
}

// GetByID retrieves a specific server master key by its ID
func (m *ServerMasterKeyModel) GetByID(keyID string) (*ServerMasterKey, error) {
	var key ServerMasterKey
	if err := m.db.Where("key_id = ?", keyID).First(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to get server key: %w", err)
	}
	return &key, nil
}

// ListAll retrieves all server master keys with their status
func (m *ServerMasterKeyModel) ListAll() ([]ServerMasterKey, error) {
	var keys []ServerMasterKey
	if err := m.db.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("failed to list server keys: %w", err)
	}
	return keys, nil
}
