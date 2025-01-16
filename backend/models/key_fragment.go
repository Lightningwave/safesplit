package models

import (
	"encoding/hex"
	"fmt"
	"log"
	"safesplit/services"
	"safesplit/utils"
	"sort"
	"time"

	"gorm.io/gorm"
)

type HolderType string

const (
	UserHolder   HolderType = "user"
	ServerHolder HolderType = "server"
)

type KeyFragment struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	FileID            uint       `json:"file_id"`
	FragmentIndex     int        `json:"fragment_index"`
	EncryptedFragment []byte     `json:"encrypted_fragment" gorm:"type:mediumblob"`
	EncryptionNonce   []byte     `json:"encryption_nonce" gorm:"type:binary(16)"`
	HolderType        HolderType `json:"holder_type" gorm:"type:enum('user','server')"`
	MasterKeyVersion  *int       `json:"master_key_version"`
	ServerKeyID       *string    `json:"server_key_id"`
	CreatedAt         time.Time  `json:"created_at"`
}

type KeyFragmentModel struct {
	db *gorm.DB
}

func NewKeyFragmentModel(db *gorm.DB) *KeyFragmentModel {
	return &KeyFragmentModel{db: db}
}

// SaveKeyFragments stores fragments with proper distribution and encryption details
func (m *KeyFragmentModel) SaveKeyFragments(tx *gorm.DB, fileID uint, shares []services.KeyShare, userID uint, serverKeyModel *ServerMasterKeyModel) error {
	// Get server key
	serverKey, err := serverKeyModel.GetActive()
	if err != nil {
		return fmt.Errorf("failed to get server key: %w", err)
	}

	decryptedServerKey, err := serverKeyModel.GetServerKey(serverKey.KeyID)
	if err != nil {
		return fmt.Errorf("failed to get decrypted server key: %w", err)
	}

	log.Printf("Number of shares to save: %d", len(shares))
	for i, share := range shares {
		log.Printf("Share %d: Index=%d, Value=%s, Length=%d bytes",
			i, share.Index, share.Value, len(share.Value))
	}

	// Get file info
	var file struct {
		Threshold        uint
		MasterKeyVersion int
		ServerKeyID      string
	}
	if err := tx.Table("files").
		Select("threshold, master_key_version, server_key_id").
		Where("id = ?", fileID).
		Scan(&file).Error; err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Server stores threshold-1 fragments
	serverFragmentCount := int(file.Threshold) - 1
	fragments := make([]KeyFragment, len(shares))

	log.Printf("Server will store %d fragments", serverFragmentCount)

	for i, share := range shares {
		isServerFragment := i < serverFragmentCount
		holderType := ServerHolder
		if !isServerFragment {
			holderType = UserHolder
		}

		// Generate nonce
		nonce, err := utils.GenerateNonce()
		if err != nil {
			return fmt.Errorf("failed to generate nonce for fragment %d: %w", i, err)
		}

		// Convert share to bytes without truncating
		shareBytes, err := hex.DecodeString(share.Value)
		if err != nil {
			return fmt.Errorf("failed to decode share value: %w", err)
		}

		log.Printf("Fragment %d: Type=%s, Share bytes=%x (length=%d)",
			i, holderType, shareBytes, len(shareBytes))

		// Encrypt fragment
		var encryptedFragment []byte

		if isServerFragment {
			log.Printf("Using server key (length: %d) to encrypt fragment %d",
				len(decryptedServerKey), i)
			encryptedFragment, err = services.EncryptMasterKey(shareBytes, decryptedServerKey, nonce)
		} else {
			var user User
			if err := tx.First(&user, userID).Error; err != nil {
				return fmt.Errorf("failed to get user info: %w", err)
			}

			log.Printf("User master key length: %d bytes", len(user.EncryptedMasterKey))
			log.Printf("User master key version: %d", user.MasterKeyVersion)

			// Take first 32 bytes of user's master key
			userKey := user.EncryptedMasterKey[:32]
			log.Printf("Using truncated user master key (length: %d) to encrypt fragment %d",
				len(userKey), i)

			encryptedFragment, err = services.EncryptMasterKey(shareBytes, userKey, nonce)
		}

		if err != nil {
			return fmt.Errorf("failed to encrypt fragment %d: %w", i, err)
		}

		fragment := KeyFragment{
			FileID:            fileID,
			FragmentIndex:     share.Index,
			EncryptedFragment: encryptedFragment,
			EncryptionNonce:   nonce,
			HolderType:        holderType,
		}

		if isServerFragment {
			fragment.ServerKeyID = &serverKey.KeyID
			fragment.MasterKeyVersion = nil
		} else {
			fragment.ServerKeyID = nil
			fragment.MasterKeyVersion = &file.MasterKeyVersion
		}

		fragments[i] = fragment
		log.Printf("Created fragment %d - Index: %d, Type: %s, Length: %d bytes, Value=%x",
			i, share.Index, holderType, len(encryptedFragment), encryptedFragment)
	}

	// Save fragments
	result := tx.Create(&fragments)
	if result.Error != nil {
		return fmt.Errorf("failed to save key fragments: %w", result.Error)
	}
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
	serverCount := 0
	userCount := 0
	fragmentSizes := make(map[int]int)
	for _, f := range fragments {
		if f.HolderType == ServerHolder {
			serverCount++
		} else {
			userCount++
		}
		fragmentSizes[f.FragmentIndex] = len(f.EncryptedFragment)
		log.Printf("Retrieved fragment - Index: %d, Type: %s, Length: %d",
			f.FragmentIndex, f.HolderType, len(f.EncryptedFragment))
	}

	log.Printf("Retrieved %d total fragments (Server: %d, User: %d) for file %d",
		len(fragments), serverCount, userCount, fileID)

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

// GetServerFragmentsForFile retrieves only server-held fragments
func (m *KeyFragmentModel) GetServerFragmentsForFile(fileID uint) ([]KeyFragment, error) {
	log.Printf("Retrieving server fragments for file %d", fileID)

	fragments, err := m.GetFragmentsByType(fileID, ServerHolder)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server fragments: %w", err)
	}

	log.Printf("Retrieved %d server fragments for file %d", len(fragments), fileID)
	return fragments, nil
}
