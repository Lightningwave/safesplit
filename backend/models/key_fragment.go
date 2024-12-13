// models/key_fragment.go
package models

import (
	"fmt"
	"time"

	"safesplit/services"

	"gorm.io/gorm"
)

type KeyFragment struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	FileID            uint      `json:"file_id"`
	FragmentIndex     int       `json:"fragment_index"`
	EncryptedFragment string    `json:"encrypted_fragment"`
	HolderType        string    `json:"holder_type" gorm:"type:varchar(50)"`
	CreatedAt         time.Time `json:"created_at"`
}

type KeyFragmentModel struct {
	db *gorm.DB
}

func NewKeyFragmentModel(db *gorm.DB) *KeyFragmentModel {
	return &KeyFragmentModel{db: db}
}

func (m *KeyFragmentModel) SaveKeyFragments(tx *gorm.DB, fileID uint, shares []services.KeyShare) error {
	fragments := make([]KeyFragment, len(shares))
	for i, share := range shares {
		fragments[i] = KeyFragment{
			FileID:            fileID,
			FragmentIndex:     share.Index,
			EncryptedFragment: share.Value,
			HolderType:        fmt.Sprintf("share-%d", share.Index),
		}
	}

	if err := tx.Create(&fragments).Error; err != nil {
		return fmt.Errorf("failed to save key fragments: %w", err)
	}

	return nil
}

func (m *KeyFragmentModel) GetKeyFragments(fileID uint) ([]KeyFragment, error) {
	var fragments []KeyFragment
	err := m.db.Where("file_id = ?", fileID).Find(&fragments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key fragments: %w", err)
	}
	return fragments, nil
}
