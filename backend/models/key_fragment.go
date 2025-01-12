package models

import (
	"fmt"
	"log"
	"safesplit/services"
	"sort"
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

// SaveKeyFragments stores fragments with proper distribution between system and user
func (m *KeyFragmentModel) SaveKeyFragments(tx *gorm.DB, fileID uint, shares []services.KeyShare) error {
	// Get file threshold
	var threshold uint
	if err := tx.Table("files").
		Select("threshold").
		Where("id = ?", fileID).
		Scan(&threshold).Error; err != nil {
		return fmt.Errorf("failed to get file threshold: %w", err)
	}

	log.Printf("Saving key fragments for file %d - Total shares: %d, Threshold: %d",
		fileID, len(shares), threshold)

	// Validate share count
	if len(shares) < int(threshold) {
		return fmt.Errorf("insufficient shares: have %d, need %d", len(shares), threshold)
	}

	// System stores threshold-1 fragments
	systemFragmentCount := int(threshold) - 1
	fragments := make([]KeyFragment, len(shares))

	for i, share := range shares {
		holderType := UserHolder
		if i < systemFragmentCount {
			holderType = SystemHolder
		}

		fragments[i] = KeyFragment{
			FileID:            fileID,
			FragmentIndex:     share.Index,
			EncryptedFragment: []byte(share.Value),
			HolderType:        holderType,
		}
		log.Printf("Creating fragment %d - Index: %d, Type: %s, Length: %d",
			i, share.Index, holderType, len(share.Value))
	}

	// Save fragments in transaction
	result := tx.Create(&fragments)
	if result.Error != nil {
		return fmt.Errorf("failed to save key fragments: %w", result.Error)
	}
	if result.RowsAffected != int64(len(fragments)) {
		return fmt.Errorf("failed to save all fragments: saved %d of %d",
			result.RowsAffected, len(fragments))
	}

	log.Printf("Successfully saved %d key fragments for file %d", len(fragments), fileID)
	return nil
}

// GetKeyFragments retrieves and verifies fragments for reconstruction
func (m *KeyFragmentModel) GetKeyFragments(fileID uint) ([]KeyFragment, error) {
	var fragments []KeyFragment

	// Get file information
	var file struct {
		Threshold  uint
		ShareCount uint
	}
	if err := m.db.Table("files").
		Select("threshold, share_count").
		Where("id = ?", fileID).
		Scan(&file).Error; err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	log.Printf("Retrieving key fragments for file %d - Threshold: %d, Expected shares: %d",
		fileID, file.Threshold, file.ShareCount)

	// Retrieve all fragments
	err := m.db.Where("file_id = ?", fileID).
		Order("fragment_index asc").
		Find(&fragments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key fragments: %w", err)
	}

	if len(fragments) == 0 {
		return nil, fmt.Errorf("no key fragments found for file ID %d", fileID)
	}

	// Count fragments by type
	systemCount := 0
	userCount := 0
	fragmentSizes := make(map[int]int)
	for _, f := range fragments {
		if f.HolderType == SystemHolder {
			systemCount++
		} else {
			userCount++
		}
		fragmentSizes[f.FragmentIndex] = len(f.EncryptedFragment)
		log.Printf("Retrieved fragment - Index: %d, Type: %s, Length: %d",
			f.FragmentIndex, f.HolderType, len(f.EncryptedFragment))
	}

	log.Printf("Retrieved %d total fragments (System: %d, User: %d) for file %d",
		len(fragments), systemCount, userCount, fileID)

	// Validate fragment count
	if len(fragments) < int(file.Threshold) {
		return nil, fmt.Errorf("insufficient fragments: have %d, need %d",
			len(fragments), file.Threshold)
	}

	// Verify fragment integrity
	expectedSize := -1
	for idx, size := range fragmentSizes {
		if expectedSize == -1 {
			expectedSize = size
		} else if size != expectedSize {
			log.Printf("Fragment size mismatch - Index %d: %d bytes (expected %d)",
				idx, size, expectedSize)
			return nil, fmt.Errorf("inconsistent fragment sizes detected")
		}
	}

	// Sort fragments by index
	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].FragmentIndex < fragments[j].FragmentIndex
	})

	log.Printf("Returning %d validated fragments for file %d", len(fragments), fileID)
	return fragments, nil
}

// GetFragmentsByType retrieves fragments of a specific type
func (m *KeyFragmentModel) GetFragmentsByType(fileID uint, holderType HolderType) ([]KeyFragment, error) {
	var fragments []KeyFragment

	log.Printf("Retrieving %s fragments for file %d", holderType, fileID)

	err := m.db.Where("file_id = ? AND holder_type = ?",
		fileID, holderType).
		Order("fragment_index asc").
		Find(&fragments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve %s fragments: %w", holderType, err)
	}

	log.Printf("Found %d fragments of type %s for file %d",
		len(fragments), holderType, fileID)

	return fragments, nil
}

// GetUserFragmentsForFile retrieves only user-held fragments
func (m *KeyFragmentModel) GetUserFragmentsForFile(fileID uint) ([]KeyFragment, error) {
	log.Printf("Retrieving user fragments for file %d", fileID)

	fragments, err := m.GetFragmentsByType(fileID, UserHolder)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user fragments: %w", err)
	}

	log.Printf("Retrieved %d user fragments for file %d", len(fragments), fileID)
	return fragments, nil
}
