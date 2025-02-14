package models

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"safesplit/services"
	"strings"
	"time"

	"gorm.io/gorm"
)

type EncryptionType string

const (
	StandardEncryption EncryptionType = "standard" // AES-256-GCM
	ChaCha20           EncryptionType = "chacha20" // ChaCha20-Poly1305
	Twofish            EncryptionType = "twofish"  // Twofish
)

type File struct {
	ID                uint                    `json:"id" gorm:"primaryKey"`
	UserID            uint                    `json:"user_id"`
	FolderID          *uint                   `json:"folder_id,omitempty"`
	Name              string                  `json:"name"`
	OriginalName      string                  `json:"original_name"`
	FilePath          string                  `json:"file_path"`
	Size              int64                   `json:"size" gorm:"not null"`
	CompressedSize    int64                   `json:"compressed_size"`
	IsCompressed      bool                    `json:"is_compressed" gorm:"default:false"`
	CompressionRatio  float64                 `json:"compression_ratio"`
	MimeType          string                  `json:"mime_type"`
	IsArchived        bool                    `json:"is_archived" gorm:"default:false"`
	IsDeleted         bool                    `json:"is_deleted" gorm:"default:false"`
	DeletedAt         *time.Time              `json:"deleted_at"`
	EncryptionIV      []byte                  `json:"encryption_iv" gorm:"type:varbinary(24);null"`
	EncryptionSalt    []byte                  `json:"encryption_salt" gorm:"type:binary(32);null"`
	EncryptionType    services.EncryptionType `json:"encryption_type" gorm:"type:varchar(20);default:'standard'"`
	EncryptionVersion int                     `json:"encryption_version" gorm:"default:1"`
	ServerKeyID       string                  `json:"server_key_id" gorm:"type:varchar(64)"`
	MasterKeyVersion  int                     `json:"master_key_version" gorm:"not null;default:1"`
	FileHash          string                  `json:"file_hash"`
	ShareCount        uint                    `json:"share_count" gorm:"not null;default:2"`
	Threshold         uint                    `json:"threshold" gorm:"not null;default:2"`
	DataShardCount    uint                    `json:"data_shard_count" gorm:"not null;default:4"`
	ParityShardCount  uint                    `json:"parity_shard_count" gorm:"not null;default:2"`
	IsSharded         bool                    `json:"is_sharded" gorm:"default:false"`
	IsShared          bool                    `json:"is_shared" gorm:"default:false"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
}
type FileModel struct {
	db                *gorm.DB
	rsService         *services.ReedSolomonService
	serverKeyModel    *ServerMasterKeyModel
	encryptionService *services.EncryptionService
	keyFragmentModel  *KeyFragmentModel
}

func NewFileModel(
	db *gorm.DB,
	rsService *services.ReedSolomonService,
	serverKeyModel *ServerMasterKeyModel,
	encryptionService *services.EncryptionService,
	keyFragmentModel *KeyFragmentModel,
) *FileModel {
	return &FileModel{
		db:                db,
		rsService:         rsService,
		serverKeyModel:    serverKeyModel,
		encryptionService: encryptionService,
		keyFragmentModel:  keyFragmentModel,
	}
}

// validation method for encryption type
func (f *File) ValidateEncryption() error {
	switch f.EncryptionType {
	case services.StandardEncryption, services.ChaCha20, services.Twofish:
		return nil
	default:
		return fmt.Errorf("unsupported encryption type: %s", f.EncryptionType)
	}
}

// File CRUD operations
func (m *FileModel) CreateFile(tx *gorm.DB, file *File) error {
	if file.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}

	if file.Name == "" {
		return fmt.Errorf("file name is required")
	}

	// Add encryption validation
	if err := file.ValidateEncryption(); err != nil {
		return err
	}

	// Add IV size validation
	if err := file.ValidateIVSize(); err != nil {
		return err
	}

	result := tx.Create(file)
	if result.Error != nil {
		return fmt.Errorf("failed to create file record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no file record created")
	}
	return nil
}

func (m *FileModel) CreateFileWithShards(
	file *File,
	shares []services.KeyShare,
	shards [][]byte,
	keyFragmentModel *KeyFragmentModel,
	serverKeyModel *ServerMasterKeyModel,
) error {
	return withTransactionRetry(m.db, 3, func(tx *gorm.DB) error {
		// 1. Update user storage first to lock the user row
		if err := m.UpdateUserStorage(tx, file.UserID, file.Size); err != nil {
			return fmt.Errorf("failed to update storage usage: %w", err)
		}

		// 2. Create file record
		if err := m.CreateFile(tx, file); err != nil {
			return fmt.Errorf("failed to create file record: %w", err)
		}

		// 3. Store shards
		if err := m.rsService.StoreShards(file.ID, &services.FileShards{Shards: shards}); err != nil {
			return fmt.Errorf("failed to store shards: %w", err)
		}

		// 4. Save key fragments
		if err := keyFragmentModel.SaveKeyFragments(tx, file.ID, shares, file.UserID, serverKeyModel); err != nil {
			m.rsService.DeleteShards(file.ID) // clean up
			return fmt.Errorf("failed to save key fragments: %w", err)
		}

		// 5. Log activity
		activity := &ActivityLog{
			UserID:       file.UserID,
			ActivityType: "upload",
			FileID:       &file.ID,
			Status:       "success",
			Details: fmt.Sprintf("File uploaded with %s encryption, %d shards",
				file.EncryptionType, len(shards)),
		}
		if err := tx.Create(activity).Error; err != nil {
			m.rsService.DeleteShards(file.ID)
			return fmt.Errorf("failed to log activity: %w", err)
		}

		return nil
	})
}

func withTransactionRetry(db *gorm.DB, maxRetries int, fn func(tx *gorm.DB) error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		tx := db.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		err = fn(tx)
		if err != nil {
			tx.Rollback()

			// Check if it's a deadlock (or use an error code check if your driver exposes it)
			if strings.Contains(strings.ToLower(err.Error()), "deadlock") {
				continue // retry
			}
			return err
		}

		err = tx.Commit().Error
		if err == nil {
			// Transaction committed successfully
			return nil
		}

		// If commit fails due to deadlock, retry
		if strings.Contains(strings.ToLower(err.Error()), "deadlock") {
			continue
		}
		return err
	}
	// If we exhausted retries, return the last error
	return err
}

// ReadFileShards retrieves and reconstructs the file content from shards
func (m *FileModel) ReadFileShards(file *File) ([]byte, error) {
	log.Printf("Reading file shards - ID: %d, Encryption: %s", file.ID, file.EncryptionType)

	// Get key fragments for decryption
	fragmentsWithData, err := m.keyFragmentModel.GetKeyFragments(file.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key fragments: %w", err)
	}

	// Convert FragmentData to KeyShare
	keyShares := make([]services.KeyShare, len(fragmentsWithData))
	for i, fragment := range fragmentsWithData {
		keyShares[i] = services.KeyShare{
			Index:      fragment.FragmentIndex,
			Value:      hex.EncodeToString(fragment.Data),
			HolderType: string(fragment.HolderType),
			NodeIndex:  fragment.NodeIndex,
		}
	}

	// Retrieve all available shards
	fileShards, err := m.rsService.RetrieveShards(file.ID, int(file.DataShardCount+file.ParityShardCount))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve shards: %w", err)
	}

	// Validate we have enough shards for reconstruction
	if !m.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
		return nil, fmt.Errorf("insufficient shards available for reconstruction")
	}

	// Reconstruct the original data
	reconstructed, err := m.rsService.ReconstructFile(fileShards.Shards, int(file.DataShardCount), int(file.ParityShardCount))
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct file: %w", err)
	}

	log.Printf("File reconstructed - Size: %d bytes", len(reconstructed))

	// Use the unified DecryptFileWithType method
	decrypted, err := m.encryptionService.DecryptFileWithType(
		reconstructed,
		file.EncryptionIV,
		keyShares,
		int(file.Threshold),
		file.EncryptionSalt,
		file.EncryptionType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file: %w", err)
	}

	log.Printf("File decrypted successfully - Final size: %d bytes", len(decrypted))
	return decrypted, nil
}
func (m *FileModel) GetFileEncryptionInfo(fileID uint) (*struct {
	Type      services.EncryptionType `json:"type"`
	Version   int                     `json:"version"`
	Algorithm string                  `json:"algorithm"`
}, error) {
	var file File
	err := m.db.Select("encryption_type, encryption_version").
		Where("id = ?", fileID).
		First(&file).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption info: %w", err)
	}

	var algorithm string
	switch file.EncryptionType {
	case services.StandardEncryption:
		algorithm = "AES-256-GCM"
	case services.ChaCha20:
		algorithm = "ChaCha20-Poly1305"
	case services.Twofish:
		algorithm = "Twofish-256"
	}

	return &struct {
		Type      services.EncryptionType `json:"type"`
		Version   int                     `json:"version"`
		Algorithm string                  `json:"algorithm"`
	}{
		Type:      file.EncryptionType,
		Version:   file.EncryptionVersion,
		Algorithm: algorithm,
	}, nil
}

func (m *FileModel) GetFileByID(fileID uint) (*File, error) {
	var file File
	err := m.db.Where("id = ? AND is_deleted = ? AND is_archived = ?", fileID, false, false).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found or inaccessible")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &file, nil
}

// GetFileForDownload updated to handle Reed-Solomon shards
func (m *FileModel) GetFileForDownload(fileID, userID uint) (*File, error) {
	var file File
	err := m.db.Where("id = ? AND user_id = ? AND is_deleted = ? AND is_archived = ?",
		fileID, userID, false, false).First(&file).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("File lookup failed - ID: %d, RequestingUser: %d", fileID, userID)
			return nil, fmt.Errorf("file not found or access denied")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	log.Printf("File download requested - ID: %d, Path: %s, Owner: %d, IsSharded: %v",
		file.ID, file.FilePath, file.UserID, file.IsSharded)

	return &file, nil
}

// File listing methods
func (m *FileModel) ListUserFiles(userID uint) ([]File, error) {
	var files []File
	err := m.db.Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("created_at DESC").
		Find(&files).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user files: %w", err)
	}

	return files, nil
}

func (m *FileModel) ListUserFilesInFolder(userID uint, folderID *uint) ([]File, error) {
	var files []File
	query := m.db.Where("user_id = ? AND is_deleted = ?", userID, false)

	if folderID != nil {
		query = query.Where("folder_id = ?", folderID)
	} else {
		query = query.Where("folder_id IS NULL")
	}

	err := query.Order("created_at DESC").Find(&files).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user files: %w", err)
	}

	return files, nil
}

// File operations
func (m *FileModel) DeleteFile(fileID, userID uint, ipAddress string) error {
	tx := m.db.Begin()

	log.Printf("Starting file deletion process - File ID: %d, User ID: %d", fileID, userID)

	var file File
	if err := tx.Where("id = ? AND user_id = ? AND is_deleted = ?", fileID, userID, false).First(&file).Error; err != nil {
		tx.Rollback()
		log.Printf("File not found or already deleted - ID: %d, Error: %v", fileID, err)
		return fmt.Errorf("file not found: %w", err)
	}

	log.Printf("Found file to delete - ID: %d, IsSharded: %v, Size: %d bytes",
		file.ID, file.IsSharded, file.Size)

	// Soft delete the file
	result := tx.Model(&file).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": time.Now(),
	})

	if result.Error != nil {
		tx.Rollback()
		log.Printf("Failed to soft delete file - ID: %d, Error: %v", fileID, result.Error)
		return fmt.Errorf("failed to delete file: %w", result.Error)
	}

	// Keep shards for potential recovery
	if !file.IsSharded && file.FilePath != "" {
		if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Failed to delete file content - Path: %s, Error: %v", file.FilePath, err)
			log.Printf("Continuing deletion process despite file removal error")
		}
	}

	// Update user storage
	if err := m.UpdateUserStorage(tx, userID, -file.Size); err != nil {
		tx.Rollback()
		log.Printf("Failed to update storage usage - User ID: %d, Error: %v", userID, err)
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	// Log activity
	activity := &ActivityLog{
		UserID:       userID,
		ActivityType: "delete",
		FileID:       &fileID,
		IPAddress:    ipAddress,
		Status:       "success",
		Details:      fmt.Sprintf("File marked as deleted (Sharded: %v, Size: %d bytes)", file.IsSharded, file.Size),
	}

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to log deletion activity - File ID: %d, Error: %v", fileID, err)
		return fmt.Errorf("failed to log activity: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit deletion transaction - File ID: %d, Error: %v", fileID, err)
		return fmt.Errorf("failed to complete delete operation: %w", err)
	}

	log.Printf("Successfully completed file deletion - ID: %d", fileID)
	return nil
}
func (m *FileModel) ArchiveFile(fileID, userID uint, ipAddress string) error {
	tx := m.db.Begin()

	// Archive the file
	result := tx.Model(&File{}).
		Where("id = ? AND user_id = ? AND is_archived = ?", fileID, userID, false).
		Update("is_archived", true)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to archive file: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("file not found or already archived")
	}

	// Log activity
	activity := &ActivityLog{
		UserID:       userID,
		ActivityType: "archive",
		FileID:       &fileID,
		IPAddress:    ipAddress,
		Status:       "success",
	}

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to log activity: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to complete archive operation: %w", err)
	}

	return nil
}
func (m *FileModel) UnarchiveFile(fileID, userID uint, ipAddress string) error {
	tx := m.db.Begin()

	// Unarchive the file
	result := tx.Model(&File{}).
		Where("id = ? AND user_id = ? AND is_archived = ?", fileID, userID, true).
		Update("is_archived", false)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unarchive file: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("file not found or not archived")
	}

	// Log activity
	activity := &ActivityLog{
		UserID:       userID,
		ActivityType: "unarchive",
		FileID:       &fileID,
		IPAddress:    ipAddress,
		Status:       "success",
	}

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to log activity: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to complete unarchive operation: %w", err)
	}

	return nil
}

// Storage management
func (m *FileModel) GetUserStorageInfo(userID uint) (used int64, quota int64, err error) {
	var user User
	err = m.db.Select("storage_used, storage_quota").First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, 0, fmt.Errorf("user not found: %w", err)
		}
		return 0, 0, fmt.Errorf("database error: %w", err)
	}
	return user.StorageUsed, user.StorageQuota, nil
}

func (m *FileModel) UpdateUserStorage(tx *gorm.DB, userID uint, size int64) error {
	result := tx.Model(&User{}).Where("id = ?", userID).
		Update("storage_used", gorm.Expr("storage_used + ?", size))
	if result.Error != nil {
		return fmt.Errorf("failed to update storage usage: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no storage update performed")
	}
	return nil
}

// Activity logging
func (m *FileModel) LogDownloadActivity(tx *gorm.DB, fileID, userID uint, ipAddress string) error {
	activity := ActivityLog{
		UserID:       userID,
		ActivityType: "download",
		FileID:       &fileID,
		IPAddress:    ipAddress,
		Status:       "success",
	}

	if err := tx.Create(&activity).Error; err != nil {
		return fmt.Errorf("failed to log download activity: %w", err)
	}

	return nil
}

// File content handling
func (m *FileModel) ReadFileContent(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found on server: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func (m *FileModel) ListRootFiles(userID uint) ([]File, error) {
	var files []File
	err := m.db.Where("user_id = ? AND folder_id IS NULL AND is_deleted = ?", userID, false).
		Order("created_at DESC").
		Find(&files).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch root files: %w", err)
	}
	return files, nil
}

func (m *FileModel) ListFolderFiles(userID uint, folderID uint) ([]File, error) {
	var files []File
	err := m.db.Where("user_id = ? AND folder_id = ? AND is_deleted = ?", userID, folderID, false).
		Order("created_at DESC").
		Find(&files).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch folder files: %w", err)
	}
	return files, nil
}
func (m *FileModel) ListAllUserFiles(userID uint) ([]File, error) {
	var files []File
	err := m.db.Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("created_at DESC").
		Find(&files).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user files: %w", err)
	}
	return files, nil
}
func (m *FileModel) GetRecoverableFiles(userID uint) ([]File, error) {
	var files []File
	err := m.db.Where("user_id = ? AND is_deleted = ?", userID, true).
		Order("deleted_at DESC").
		Find(&files).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch deleted files: %w", err)
	}

	return files, nil
}

// RecoverFile recovers a deleted file and updates storage usage
func (m *FileModel) RecoverFile(fileID, userID uint) error {
	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered in RecoverFile: %v", r)
			tx.Rollback()
		}
	}()

	log.Printf("Starting file recovery process - File ID: %d, User ID: %d", fileID, userID)

	var file File
	if err := tx.Where("id = ? AND user_id = ? AND is_deleted = ?", fileID, userID, true).First(&file).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to find deleted file: ID=%d, Error: %v", fileID, err)
		return fmt.Errorf("file not found or already recovered: %w", err)
	}

	log.Printf("Found deleted file - ID: %d, IsSharded: %v, Size: %d bytes",
		file.ID, file.IsSharded, file.Size)

	usedStorage, quota, err := m.GetUserStorageInfo(userID)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to get storage info - User ID: %d, Error: %v", userID, err)
		return fmt.Errorf("failed to verify storage availability: %w", err)
	}

	if usedStorage+file.Size > quota {
		tx.Rollback()
		log.Printf("Insufficient storage space - Used: %d, Quota: %d, Required: %d",
			usedStorage, quota, file.Size)
		return fmt.Errorf("insufficient storage space for recovery")
	}

	if file.IsSharded {
		log.Printf("Verifying shards for file %d", file.ID)
		fileShards, err := m.rsService.RetrieveShards(file.ID, int(file.DataShardCount+file.ParityShardCount))
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to retrieve shards: %v", err)
			return fmt.Errorf("failed to verify file shards: %w", err)
		}

		validShards := 0
		for i, shard := range fileShards.Shards {
			if shard != nil {
				validShards++
				log.Printf("Found valid shard %d: %d bytes", i, len(shard))
			} else {
				log.Printf("Missing shard %d", i)
			}
		}

		if !m.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
			tx.Rollback()
			log.Printf("Insufficient shards - Found %d, Need %d", validShards, file.DataShardCount)
			return fmt.Errorf("cannot recover file: insufficient shards available (%d/%d)",
				validShards, file.DataShardCount)
		}

		log.Printf("Verified shards availability - Found %d valid shards", validShards)
	} else {
		// Verify regular file exists
		if file.FilePath != "" {
			if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
				tx.Rollback()
				log.Printf("File content not found at path: %s", file.FilePath)
				return fmt.Errorf("file content not found: %w", err)
			}
		}
	}

	// Update file status
	updateResult := tx.Model(&file).Updates(map[string]interface{}{
		"is_deleted": false,
		"deleted_at": nil,
	})
	if updateResult.Error != nil {
		tx.Rollback()
		log.Printf("Failed to update file status: %v", updateResult.Error)
		return fmt.Errorf("failed to recover file: %w", updateResult.Error)
	}
	if updateResult.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("No rows affected when updating file status")
		return fmt.Errorf("file recovery failed: no changes made")
	}

	// Update user storage
	if err := m.UpdateUserStorage(tx, userID, file.Size); err != nil {
		tx.Rollback()
		log.Printf("Failed to update storage usage: %v", err)
		return fmt.Errorf("failed to update storage: %w", err)
	}

	// Log recovery activity
	activity := &ActivityLog{
		UserID:       userID,
		ActivityType: "restore",
		FileID:       &file.ID,
		Status:       "success",
		Details:      fmt.Sprintf("File recovered (Sharded: %v, Size: %d bytes)", file.IsSharded, file.Size),
	}

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to log recovery activity: %v", err)
		return fmt.Errorf("failed to log recovery activity: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit recovery transaction: %v", err)
		return fmt.Errorf("failed to complete recovery: %w", err)
	}

	log.Printf("Successfully recovered file - ID: %d", file.ID)
	return nil
}

// GetUserFileCount returns the total number of non-deleted files for a user
func (m *FileModel) GetUserFileCount(userID uint) (int64, error) {
	var count int64
	err := m.db.Model(&File{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count files: %w", err)
	}
	return count, nil
}
func (f *File) ValidateIVSize() error {
	var expectedSize int
	switch f.EncryptionType {
	case services.ChaCha20:
		expectedSize = 24 
	case services.Twofish:
		expectedSize = 12 
	case services.StandardEncryption:
		expectedSize = 16 
	default:
		return fmt.Errorf("unsupported encryption type: %s", f.EncryptionType)
	}

	if len(f.EncryptionIV) != expectedSize {
		return fmt.Errorf("invalid IV length for %s encryption: got %d, expected %d",
			f.EncryptionType, len(f.EncryptionIV), expectedSize)
	}
	return nil
}
func (m *FileModel) PermanentlyDeleteFile(fileID, userID uint, ipAddress string) error {
    tx := m.db.Begin()

    log.Printf("Starting permanent file deletion process - File ID: %d, User ID: %d", fileID, userID)

    // Get file info first
    var file File
    if err := tx.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
        tx.Rollback()
        log.Printf("File not found - ID: %d, Error: %v", fileID, err)
        return fmt.Errorf("file not found: %w", err)
    }

    log.Printf("Found file to permanently delete - ID: %d, IsSharded: %v, Size: %d bytes",
        file.ID, file.IsSharded, file.Size)

    // Handle physical data deletion
    if file.IsSharded {
        log.Printf("Deleting Reed-Solomon shards for file %d", fileID)
        // Delete shards
        if err := m.rsService.DeleteShards(fileID); err != nil {
            tx.Rollback()
            log.Printf("Failed to delete shards - File ID: %d, Error: %v", fileID, err)
            return fmt.Errorf("failed to delete shards: %w", err)
        }
        log.Printf("Successfully deleted shards for file %d", fileID)
    } else if file.FilePath != "" {
        // Delete regular file if it exists
        if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
            log.Printf("Failed to delete file content - Path: %s, Error: %v", file.FilePath, err)
            // Don't rollback here as we want to continue with database cleanup
        }
    }

    // Delete related key fragments
    if err := tx.Where("file_id = ?", fileID).Delete(&KeyFragment{}).Error; err != nil {
        tx.Rollback()
        log.Printf("Failed to delete key fragments - File ID: %d, Error: %v", fileID, err)
        return fmt.Errorf("failed to delete key fragments: %w", err)
    }

    // Delete related activity logs
    if err := tx.Where("file_id = ?", fileID).Delete(&ActivityLog{}).Error; err != nil {
        tx.Rollback()
        log.Printf("Failed to delete activity logs - File ID: %d, Error: %v", fileID, err)
        return fmt.Errorf("failed to delete activity logs: %w", err)
    }

    // Finally, delete the file record
    if err := tx.Delete(&file).Error; err != nil {
        tx.Rollback()
        log.Printf("Failed to delete file record - ID: %d, Error: %v", fileID, err)
        return fmt.Errorf("failed to delete file record: %w", err)
    }

    // Log the permanent deletion activity (in a separate table for audit history)
    permanentDeletionLog := &PermanentDeletionLog{
        UserID:       userID,
        FileName:     file.Name,
        OriginalID:   fileID,
        Size:        file.Size,
        IsSharded:   file.IsSharded,
        IPAddress:   ipAddress,
        DeletedAt:   time.Now(),
        Details:     fmt.Sprintf("File permanently deleted (Sharded: %v, Size: %d bytes)", file.IsSharded, file.Size),
    }

    if err := tx.Create(permanentDeletionLog).Error; err != nil {
        tx.Rollback()
        log.Printf("Failed to create permanent deletion log - File ID: %d, Error: %v", fileID, err)
        return fmt.Errorf("failed to log permanent deletion: %w", err)
    }

    if err := tx.Commit().Error; err != nil {
        log.Printf("Failed to commit permanent deletion transaction - File ID: %d, Error: %v", fileID, err)
        return fmt.Errorf("failed to complete permanent deletion: %w", err)
    }

    log.Printf("Successfully completed permanent file deletion - ID: %d", fileID)
    return nil
}

type PermanentDeletionLog struct {
    ID          uint      `gorm:"primaryKey"`
    UserID      uint      `gorm:"not null"`
    FileName    string    `gorm:"not null"`
    OriginalID  uint      `gorm:"not null"`
    Size        int64     `gorm:"not null"`
    IsSharded   bool      `gorm:"not null"`
    IPAddress   string    `gorm:"not null"`
    DeletedAt   time.Time `gorm:"not null"`
    Details     string    `gorm:"type:text"`
}
func (m *FileModel) CleanupOldDeletedFiles(retentionDays int) error {
    cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
    
    var filesToDelete []File
    if err := m.db.Where("is_deleted = ? AND deleted_at < ?", true, cutoffDate).Find(&filesToDelete).Error; err != nil {
        return fmt.Errorf("failed to fetch old deleted files: %w", err)
    }

    for _, file := range filesToDelete {
        log.Printf("Cleaning up old deleted file - ID: %d, DeletedAt: %v", file.ID, file.DeletedAt)
        if err := m.PermanentlyDeleteFile(file.ID, file.UserID, "system"); err != nil {
            log.Printf("Failed to cleanup file %d: %v", file.ID, err)
            // Continue with other files even if one fails
            continue
        }
    }

    return nil
}