package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"safesplit/services"
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	UserID           uint       `json:"user_id"`
	FolderID         *uint      `json:"folder_id,omitempty"`
	Name             string     `json:"name"`
	OriginalName     string     `json:"original_name"`
	FilePath         string     `json:"file_path"`
	Size             int64      `json:"size" gorm:"not null"`
	CompressedSize   int64      `json:"compressed_size"`
	IsCompressed     bool       `json:"is_compressed" gorm:"default:false"`
	CompressionRatio float64    `json:"compression_ratio"`
	MimeType         string     `json:"mime_type"`
	IsArchived       bool       `json:"is_archived" gorm:"default:false"`
	IsDeleted        bool       `json:"is_deleted" gorm:"default:false"`
	DeletedAt        *time.Time `json:"deleted_at"`
	EncryptionIV     []byte     `json:"encryption_iv" gorm:"type:binary(16);null"`
	EncryptionSalt   []byte     `json:"encryption_salt" gorm:"type:binary(32);null"`
	FileHash         string     `json:"file_hash"`
	ShareCount       uint       `json:"share_count" gorm:"not null;default:2"`
	Threshold        uint       `json:"threshold" gorm:"not null;default:2"`
	DataShardCount   uint       `json:"data_shard_count" gorm:"not null;default:4"`
	ParityShardCount uint       `json:"parity_shard_count" gorm:"not null;default:2"`
	IsSharded        bool       `json:"is_sharded" gorm:"default:false"`
	IsShared         bool       `json:"is_shared" gorm:"default:false"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type FileModel struct {
	db              *gorm.DB
	storageBasePath string
	rsService       *services.ReedSolomonService
}

func NewFileModel(db *gorm.DB, rsService *services.ReedSolomonService) *FileModel {
	return &FileModel{
		db:              db,
		storageBasePath: filepath.Join("storage", "files"),
		rsService:       rsService,
	}
}

// Storage management methods
func (m *FileModel) InitializeStorage() error {
	if err := os.MkdirAll(m.storageBasePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}
	return nil
}

func (m *FileModel) GenerateStoragePath(fileName string) string {
	return filepath.Join(m.storageBasePath, fileName)
}

func (m *FileModel) SaveFileContent(fileName string, content []byte) error {
	filePath := m.GenerateStoragePath(fileName)
	if err := os.WriteFile(filePath, content, 0600); err != nil {
		return fmt.Errorf("failed to save file content: %w", err)
	}
	return nil
}

func (m *FileModel) DeleteFileContent(fileName string) error {
	filePath := m.GenerateStoragePath(fileName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file content: %w", err)
	}
	return nil
}

// File CRUD operations
func (m *FileModel) CreateFile(tx *gorm.DB, file *File) error {
	if file.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}

	if file.Name == "" {
		return fmt.Errorf("file name is required")
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

func (m *FileModel) CreateFileWithShards(file *File, shares []services.KeyShare, shards [][]byte, keyFragmentModel *KeyFragmentModel) error {
	if err := m.InitializeStorage(); err != nil {
		return err
	}

	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Verify folder if provided
	if file.FolderID != nil {
		var folder Folder
		if err := tx.Where("id = ? AND user_id = ? AND is_archived = ?",
			file.FolderID, file.UserID, false).First(&folder).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("invalid folder: %w", err)
		}
	}

	// Create file record
	if err := m.CreateFile(tx, file); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create file record: %w", err)
	}

	// Store the Reed-Solomon shards
	if err := m.rsService.StoreShards(file.ID, &services.FileShards{Shards: shards}); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to store shards: %w", err)
	}

	// Save key fragments
	if err := keyFragmentModel.SaveKeyFragments(tx, file.ID, shares); err != nil {
		// Clean up stored shards on failure
		m.rsService.DeleteShards(file.ID)
		tx.Rollback()
		return fmt.Errorf("failed to save key fragments: %w", err)
	}

	// Update user storage
	if err := m.UpdateUserStorage(tx, file.UserID, file.Size); err != nil {
		// Clean up stored shards on failure
		m.rsService.DeleteShards(file.ID)
		tx.Rollback()
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		// Clean up stored shards on failure
		m.rsService.DeleteShards(file.ID)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReadFileShards retrieves and reconstructs the file content from shards
func (m *FileModel) ReadFileShards(file *File) ([]byte, error) {
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

	return reconstructed, nil
}

func (m *FileModel) GetFileByID(fileID uint) (*File, error) {
	var file File
	err := m.db.Where("id = ? AND is_deleted = ?", fileID, false).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &file, nil
}

// GetFileForDownload updated to handle Reed-Solomon shards
func (m *FileModel) GetFileForDownload(fileID, userID uint) (*File, error) {
	var file File
	err := m.db.Where("id = ? AND user_id = ? AND is_deleted = ?", fileID, userID, false).
		First(&file).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("File lookup failed - ID: %d, RequestingUser: %d", fileID, userID)
			return nil, fmt.Errorf("file not found or access denied")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Log the request
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
// DeleteFile updates to handle Reed-Solomon shards
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

	// Handle physical file deletion based on storage type
	if file.IsSharded {
		log.Printf("Deleting Reed-Solomon shards for file %d", fileID)
		// Delete shards
		if err := m.rsService.DeleteShards(fileID); err != nil {
			tx.Rollback()
			log.Printf("Failed to delete shards - File ID: %d, Error: %v", fileID, err)
			return fmt.Errorf("failed to delete shards: %w", err)
		}
		log.Printf("Successfully deleted shards for file %d", fileID)
	} else {
		// Delete regular file if it exists
		if file.FilePath != "" {
			if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
				log.Printf("Failed to delete file content - Path: %s, Error: %v", file.FilePath, err)
				// Don't rollback here as the file might have been already moved/deleted
				log.Printf("Continuing deletion process despite file removal error")
			}
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
		Details:      fmt.Sprintf("File deleted (Sharded: %v, Size: %d bytes)", file.IsSharded, file.Size),
	}

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to log deletion activity - File ID: %d, Error: %v", fileID, err)
		return fmt.Errorf("failed to log activity: %w", err)
	}

	// Commit transaction
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
// RecoverFile updated to handle Reed-Solomon shards
func (m *FileModel) RecoverFile(fileID, userID uint) error {
	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered in RecoverFile: %v", r)
			tx.Rollback()
		}
	}()

	log.Printf("Starting file recovery process - File ID: %d, User ID: %d", fileID, userID)

	// Get the deleted file
	var file File
	if err := tx.Where("id = ? AND user_id = ? AND is_deleted = ?", fileID, userID, true).First(&file).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to find deleted file: ID=%d, Error: %v", fileID, err)
		return fmt.Errorf("file not found or already recovered: %w", err)
	}

	log.Printf("Found deleted file - ID: %d, IsSharded: %v, Size: %d bytes",
		file.ID, file.IsSharded, file.Size)

	// Verify storage space
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

	// Verify file data
	if file.IsSharded {
		log.Printf("Verifying shards for file %d", file.ID)
		// Check shards integrity
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
		ActivityType: "recover",
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
