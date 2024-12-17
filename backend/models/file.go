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
	IsShared         bool       `json:"is_shared" gorm:"default:false"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type FileModel struct {
	db              *gorm.DB
	storageBasePath string
}

func NewFileModel(db *gorm.DB) *FileModel {
	return &FileModel{
		db:              db,
		storageBasePath: filepath.Join("storage", "files"),
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

func (m *FileModel) CreateFileWithFragments(file *File, shares []services.KeyShare, keyFragmentModel *KeyFragmentModel) error {
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

	// Set storage path
	file.FilePath = m.GenerateStoragePath(file.Name)

	// Create file record
	if err := m.CreateFile(tx, file); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create file record: %w", err)
	}

	// Save key fragments
	if err := keyFragmentModel.SaveKeyFragments(tx, file.ID, shares); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save key fragments: %w", err)
	}

	// Update user storage
	if err := m.UpdateUserStorage(tx, file.UserID, file.Size); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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

	log.Printf("File download requested - ID: %d, Path: %s, Owner: %d", file.ID, file.FilePath, file.UserID)
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

	// Soft delete the file
	result := tx.Model(&File{}).
		Where("id = ? AND user_id = ? AND is_deleted = ?", fileID, userID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": time.Now(),
		})

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete file: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("file not found or already deleted")
	}

	// Log activity
	activity := &ActivityLog{
		UserID:       userID,
		ActivityType: "delete",
		FileID:       &fileID,
		IPAddress:    ipAddress,
		Status:       "success",
	}

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to log activity: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to complete delete operation: %w", err)
	}

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
