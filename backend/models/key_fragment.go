package models

import (
	"fmt"
	"safesplit/services"
	"time"

	"gorm.io/gorm"
)

type KeyFragment struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	FileID            uint      `json:"file_id"`
	FragmentIndex     int       `json:"fragment_index"`
	EncryptedFragment []byte    `json:"encrypted_fragment" gorm:"type:mediumblob"`
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
		// Convert the share value string to []byte
		encryptedFragment := []byte(share.Value)

		fragments[i] = KeyFragment{
			FileID:            fileID,
			FragmentIndex:     share.Index,
			EncryptedFragment: encryptedFragment,
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
