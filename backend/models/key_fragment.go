package models

import (
	"encoding/hex"
	"fmt"
	"log"
	"safesplit/services"
	"safesplit/utils"

	"gorm.io/gorm"
)

type HolderType string

const (
	UserHolder   HolderType = "user"
	ServerHolder HolderType = "server"
)

// KeyFragment represents the database model for key fragment metadata
type KeyFragment struct {
	ID               uint       `gorm:"primaryKey"`
	FileID           uint       `gorm:"not null"`
	FragmentIndex    int        `gorm:"not null"`
	FragmentPath     string     `gorm:"not null"`
	NodeIndex        int        `gorm:"not null"`
	EncryptionNonce  []byte     `gorm:"type:binary(16);not null"`
	HolderType       HolderType `gorm:"type:enum('user','server');not null"`
	MasterKeyVersion *int
	ServerKeyID      *string
}

// FragmentData represents a fragment with its data loaded from node storage
type FragmentData struct {
	KeyFragment
	Data []byte
}

type KeyFragmentModel struct {
	db      *gorm.DB
	storage *services.DistributedStorageService
}

func NewKeyFragmentModel(db *gorm.DB, storage *services.DistributedStorageService) *KeyFragmentModel {
	return &KeyFragmentModel{
		db:      db,
		storage: storage,
	}
}

func (m *KeyFragmentModel) GetKeyFragments(fileID uint) ([]FragmentData, error) {
	var fragments []KeyFragment

	// Get metadata from the database
	if err := m.db.Where("file_id = ?", fileID).
		Order("fragment_index asc").
		Find(&fragments).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve fragment metadata: %w", err)
	}

	// We'll store only successfully retrieved fragments here
	var fragmentsWithData []FragmentData

	// Attempt to load each fragment from node storage
	for _, fragment := range fragments {
		data, err := m.storage.RetrieveFragment(fragment.NodeIndex, fragment.FragmentPath)
		if err != nil {
			// Instead of returning an error, just log and skip
			log.Printf(
				"Warning: skipping missing fragment (file_id=%d, index=%d, node=%d, path=%s): %v",
				fragment.FileID, fragment.FragmentIndex, fragment.NodeIndex, fragment.FragmentPath, err,
			)
			continue
		}

		fragmentsWithData = append(fragmentsWithData, FragmentData{
			KeyFragment: fragment,
			Data:        data,
		})
	}

	log.Printf(
		"Retrieved %d fragments (of %d total) for file %d",
		len(fragmentsWithData), len(fragments), fileID,
	)
	return fragmentsWithData, nil
}

func (m *KeyFragmentModel) SaveKeyFragments(tx *gorm.DB, fileID uint, shares []services.KeyShare, userID uint, serverKeyModel *ServerMasterKeyModel) error {
	// Get server key for server fragments
	serverKey, err := serverKeyModel.GetActive()
	if err != nil {
		return fmt.Errorf("failed to get server key: %w", err)
	}

	decryptedServerKey, err := serverKeyModel.GetServerKey(serverKey.KeyID)
	if err != nil {
		return fmt.Errorf("failed to get decrypted server key: %w", err)
	}

	// Get user for user fragments
	var user User
	if err := tx.First(&user, userID).Error; err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Use first 32 bytes of user's master key
	userMasterKey := user.EncryptedMasterKey[:32]

	log.Printf("Number of shares to save: %d", len(shares))
	for i, share := range shares {
		log.Printf("Share %d: Index=%d, Value=%s, Length=%d bytes",
			i, share.Index, share.Value, len(share.Value))
	}

	serverFragmentCount := (len(shares) + 1) / 2
	fragments := make([]KeyFragment, len(shares))

	log.Printf("Server will store %d fragments", serverFragmentCount)

	for i, share := range shares {
		isServerFragment := i < serverFragmentCount
		holderType := ServerHolder
		if !isServerFragment {
			holderType = UserHolder
		}

		nonce, err := utils.GenerateNonce()
		if err != nil {
			return fmt.Errorf("failed to generate nonce for fragment %d: %w", i, err)
		}

		shareBytes, err := hex.DecodeString(share.Value)
		if err != nil {
			return fmt.Errorf("failed to decode share value: %w", err)
		}

		log.Printf("Fragment %d: Type=%s, Share bytes=%x (length=%d)",
			i, holderType, shareBytes, len(shareBytes))

		var encryptedFragment []byte
		var masterKeyVersion *int
		var serverKeyID *string

		if isServerFragment {
			log.Printf("Using server key (length: %d) to encrypt fragment %d",
				len(decryptedServerKey), i)
			encryptedFragment, err = services.EncryptMasterKey(shareBytes, decryptedServerKey, nonce)
			serverKeyID = &serverKey.KeyID
		} else {
			log.Printf("Using user master key to encrypt fragment %d", i)
			version := user.MasterKeyVersion
			masterKeyVersion = &version
			encryptedFragment, err = services.EncryptMasterKey(shareBytes, userMasterKey, nonce)
		}

		if err != nil {
			return fmt.Errorf("failed to encrypt fragment %d: %w", i, err)
		}

		// Store fragment in node
		nodeIndex := i % m.storage.NodeCount()
		fragmentPath := fmt.Sprintf("file_%d/fragment_%d", fileID, share.Index)

		if err := m.storage.StoreFragment(nodeIndex, fragmentPath, encryptedFragment); err != nil {
			return fmt.Errorf("failed to store fragment in node: %w", err)
		}

		// Create database record
		fragments[i] = KeyFragment{
			FileID:           fileID,
			FragmentIndex:    share.Index,
			FragmentPath:     fragmentPath,
			NodeIndex:        nodeIndex,
			EncryptionNonce:  nonce,
			HolderType:       holderType,
			MasterKeyVersion: masterKeyVersion,
			ServerKeyID:      serverKeyID,
		}

		log.Printf("Created fragment %d - Index: %d, Type: %s, Node: %d",
			i, share.Index, holderType, nodeIndex)
	}

	// Save metadata to database
	if err := tx.Create(&fragments).Error; err != nil {
		return fmt.Errorf("failed to save fragment metadata: %w", err)
	}

	return nil
}

func (m *KeyFragmentModel) GetFragmentsByType(fileID uint, holderType HolderType) ([]FragmentData, error) {
	var fragments []KeyFragment

	if err := m.db.Where("file_id = ? AND holder_type = ?", fileID, holderType).
		Order("fragment_index asc").
		Find(&fragments).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve fragments: %w", err)
	}

	var fragmentsWithData []FragmentData
	for _, fragment := range fragments {
		data, err := m.storage.RetrieveFragment(fragment.NodeIndex, fragment.FragmentPath)
		if err != nil {
			// If you want to skip missing, do so here
			log.Printf("Warning: skipping missing %s fragment (file_id=%d, index=%d): %v",
				holderType, fileID, fragment.FragmentIndex, err)
			continue
		}
		fragmentsWithData = append(fragmentsWithData, FragmentData{
			KeyFragment: fragment,
			Data:        data,
		})
	}

	return fragmentsWithData, nil
}

func (m *KeyFragmentModel) GetUserFragmentsForFile(fileID uint) ([]FragmentData, error) {
	return m.GetFragmentsByType(fileID, UserHolder)
}

func (m *KeyFragmentModel) GetServerFragmentsForFile(fileID uint) ([]FragmentData, error) {
	return m.GetFragmentsByType(fileID, ServerHolder)
}

// DeleteFragments removes both metadata and stored fragments
func (m *KeyFragmentModel) DeleteFragments(fileID uint) error {
	var fragments []KeyFragment
	if err := m.db.Where("file_id = ?", fileID).Find(&fragments).Error; err != nil {
		return fmt.Errorf("failed to find fragments to delete: %w", err)
	}

	// Delete from nodes
	for _, fragment := range fragments {
		if err := m.storage.DeleteFragment(fragment.NodeIndex, fragment.FragmentPath); err != nil {
			log.Printf("Warning: failed to delete fragment from node: %v", err)
		}
	}

	// Delete from database
	if err := m.db.Where("file_id = ?", fileID).Delete(&KeyFragment{}).Error; err != nil {
		return fmt.Errorf("failed to delete fragment metadata: %w", err)
	}

	return nil
}
