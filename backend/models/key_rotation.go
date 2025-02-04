package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type RotationType string

const (
    RotationTypeAutomatic RotationType = "automatic"
    RotationTypeManual    RotationType = "manual"
    RotationTypeForced    RotationType = "forced"
)
type KeyRotationHistory struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	UserID        uint         `json:"user_id"`
	OldKeyVersion int          `json:"old_key_version"`
	NewKeyVersion int          `json:"new_key_version"`
	RotationType  RotationType `json:"rotation_type" gorm:"type:enum('automatic','manual','forced','password_change')"`
	RotatedAt     time.Time    `json:"rotated_at" gorm:"autoCreateTime"`
}

type KeyRotationModel struct {
	db *gorm.DB
}

func NewKeyRotationModel(db *gorm.DB) *KeyRotationModel {
	return &KeyRotationModel{db: db}
}

// LogRotation records a key rotation event
func (m *KeyRotationModel) LogRotation(userID uint, oldVersion, newVersion int, rotationType RotationType) error {
	rotation := KeyRotationHistory{
		UserID:        userID,
		OldKeyVersion: oldVersion,
		NewKeyVersion: newVersion,
		RotationType:  rotationType,
	}

	if err := m.db.Create(&rotation).Error; err != nil {
		return fmt.Errorf("failed to log key rotation: %w", err)
	}

	return nil
}

// GetRotationHistory retrieves all rotation events for a user
func (m *KeyRotationModel) GetRotationHistory(userID uint) ([]KeyRotationHistory, error) {
	var history []KeyRotationHistory

	err := m.db.Where("user_id = ?", userID).
		Order("rotated_at DESC").
		Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve rotation history: %w", err)
	}

	return history, nil
}

// GetLatestRotation gets the most recent key rotation event for a user
func (m *KeyRotationModel) GetLatestRotation(userID uint) (*KeyRotationHistory, error) {
	var rotation KeyRotationHistory

	err := m.db.Where("user_id = ?", userID).
		Order("rotated_at DESC").
		First(&rotation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve latest rotation: %w", err)
	}

	return &rotation, nil
}

// CountRotationsByType counts rotations by type within a time period
func (m *KeyRotationModel) CountRotationsByType(userID uint, rotationType RotationType, since time.Time) (int64, error) {
	var count int64

	err := m.db.Model(&KeyRotationHistory{}).
		Where("user_id = ? AND rotation_type = ? AND rotated_at >= ?", userID, rotationType, since).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count rotations: %w", err)
	}

	return count, nil
}

// CheckRotationNeeded determines if a key rotation is needed based on time since last rotation
func (m *KeyRotationModel) CheckRotationNeeded(userID uint, maxAge time.Duration) (bool, error) {
	lastRotation, err := m.GetLatestRotation(userID)
	if err != nil {
		return false, err
	}

	// If no rotation history exists, rotation is needed
	if lastRotation == nil {
		return true, nil
	}

	// Calculate time since last rotation
	timeSinceRotation := time.Since(lastRotation.RotatedAt)
	return timeSinceRotation >= maxAge, nil
}

// GetRotationsBetween gets all rotations between two timestamps
func (m *KeyRotationModel) GetRotationsBetween(userID uint, start, end time.Time) ([]KeyRotationHistory, error) {
	var rotations []KeyRotationHistory

	err := m.db.Where("user_id = ? AND rotated_at BETWEEN ? AND ?", userID, start, end).
		Order("rotated_at DESC").
		Find(&rotations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve rotations: %w", err)
	}

	return rotations, nil
}

// GetUsersByRotationStatus gets users who need key rotation based on age threshold
func (m *KeyRotationModel) GetUsersByRotationStatus(maxAge time.Duration) ([]uint, error) {
	var userIDs []uint
	threshold := time.Now().Add(-maxAge)

	// Subquery to get latest rotation per user
	latestRotations := m.db.Table("key_rotation_history").
		Select("user_id, MAX(rotated_at) as last_rotation").
		Group("user_id")

	// Get users whose last rotation is older than threshold or who have no rotations
	err := m.db.Table("users").
		Select("users.id").
		Joins("LEFT JOIN (?) as latest_rotations ON users.id = latest_rotations.user_id", latestRotations).
		Where("latest_rotations.last_rotation < ? OR latest_rotations.last_rotation IS NULL", threshold).
		Pluck("users.id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get users needing rotation: %w", err)
	}

	return userIDs, nil
}

// ValidateKeyVersion checks if a given key version matches the latest version for a user
func (m *KeyRotationModel) ValidateKeyVersion(userID uint, version int) (bool, error) {
	latestRotation, err := m.GetLatestRotation(userID)
	if err != nil {
		return false, err
	}

	if latestRotation == nil {
		// If no rotation history, version should be 1
		return version == 1, nil
	}

	return version == latestRotation.NewKeyVersion, nil
}
