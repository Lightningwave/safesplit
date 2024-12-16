package models

import (
	"fmt"
	"safesplit/services"
	"time"

	"gorm.io/gorm"
)

type HolderType string

const (
	UserHolder   HolderType = "user"
	SystemHolder HolderType = "system"
)

type KeyFragment struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	FileID            uint       `json:"file_id"`
	FragmentIndex     int        `json:"fragment_index"`
	EncryptedFragment []byte     `json:"encrypted_fragment" gorm:"type:mediumblob"`
	HolderType        HolderType `json:"holder_type" gorm:"type:varchar(50)"`
	CreatedAt         time.Time  `json:"created_at"`
}

type KeyFragmentModel struct {
	db *gorm.DB
}

func NewKeyFragmentModel(db *gorm.DB) *KeyFragmentModel {
	return &KeyFragmentModel{db: db}
}

// SaveKeyFragments stores threshold-1 fragments in the system and the rest with the user
func (m *KeyFragmentModel) SaveKeyFragments(tx *gorm.DB, fileID uint, shares []services.KeyShare) error {
	// Get just the threshold value directly
	var threshold uint
	if err := tx.Table("files").
		Select("threshold").
		Where("id = ?", fileID).
		Scan(&threshold).Error; err != nil {
		return fmt.Errorf("failed to get file threshold: %w", err)
	}

	// System stores threshold-1 fragments
	systemFragmentCount := int(threshold) - 1
	fragments := make([]KeyFragment, len(shares))

	for i, share := range shares {
		holderType := UserHolder
		// First threshold-1 fragments go to system
		if i < systemFragmentCount {
			holderType = SystemHolder
		}

		fragments[i] = KeyFragment{
			FileID:            fileID,
			FragmentIndex:     share.Index,
			EncryptedFragment: []byte(share.Value),
			HolderType:        holderType,
		}
	}

	if err := tx.Create(&fragments).Error; err != nil {
		return fmt.Errorf("failed to save key fragments: %w", err)
	}

	return nil
}

func (m *KeyFragmentModel) GetKeyFragments(fileID uint) ([]KeyFragment, error) {
	var fragments []KeyFragment
	err := m.db.Where("file_id = ?", fileID).
		Order("fragment_index asc").
		Find(&fragments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key fragments: %w", err)
	}

	if len(fragments) == 0 {
		return nil, fmt.Errorf("no key fragments found for file ID %d", fileID)
	}

	return fragments, nil
}

func (m *KeyFragmentModel) GetFragmentsByType(fileID uint, holderType HolderType) ([]KeyFragment, error) {
	var fragments []KeyFragment
	err := m.db.Where("file_id = ? AND holder_type = ?",
		fileID, holderType).
		Order("fragment_index asc").
		Find(&fragments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key fragments: %w", err)
	}

	return fragments, nil
}
