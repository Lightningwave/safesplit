package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	UserID         uint       `json:"user_id"`
	FolderID       *uint      `json:"folder_id,omitempty"`
	Name           string     `json:"name"`
	OriginalName   string     `json:"original_name"`
	FilePath       string     `json:"file_path"`
	Size           int64      `json:"size"`
	MimeType       string     `json:"mime_type"`
	IsArchived     bool       `json:"is_archived" gorm:"default:false"`
	IsDeleted      bool       `json:"is_deleted" gorm:"default:false"`
	DeletedAt      *time.Time `json:"deleted_at"`
	EncryptionIV   []byte     `json:"encryption_iv" gorm:"type:binary(16);null"`
	EncryptionSalt []byte     `json:"encryption_salt" gorm:"type:binary(32);null"`
	FileHash       string     `json:"file_hash"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type FileModel struct {
	db *gorm.DB
}

func NewFileModel(db *gorm.DB) *FileModel {
	return &FileModel{db: db}
}

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
