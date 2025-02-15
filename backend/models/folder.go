package models

import (
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type Folder struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id"`
	Name           string    `json:"name"`
	ParentFolderID *uint     `json:"parent_folder_id,omitempty"`
	IsArchived     bool      `json:"is_archived" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Files          []File    `json:"files,omitempty" gorm:"foreignKey:FolderID"`
	SubFolders     []Folder  `json:"sub_folders,omitempty" gorm:"foreignKey:ParentFolderID"`
}

type FolderModel struct {
	db *gorm.DB
}

func NewFolderModel(db *gorm.DB) *FolderModel {
	return &FolderModel{db: db}
}

// GetUserFolders gets all root folders for a user
func (m *FolderModel) GetUserFolders(userID uint) ([]Folder, error) {
	var folders []Folder
	err := m.db.Where("user_id = ? AND parent_folder_id IS NULL AND is_archived = ?", userID, false).
		Order("name ASC").
		Find(&folders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user folders: %w", err)
	}
	return folders, nil
}

// GetFolderContents gets a folder's contents including files and subfolders
func (m *FolderModel) GetFolderContents(folderID, userID uint) (*Folder, error) {
	var folder Folder
	err := m.db.Where("id = ? AND user_id = ? AND is_archived = ?", folderID, userID, false).
		Preload("Files", "is_deleted = ?", false).
		Preload("SubFolders", "is_archived = ?", false).
		First(&folder).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to fetch folder contents: %w", err)
	}
	return &folder, nil
}

// CreateFolder creates a new folder
func (m *FolderModel) CreateFolder(folder *Folder) error {
	if folder.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}
	if folder.Name == "" {
		return fmt.Errorf("folder name is required")
	}

	if folder.ParentFolderID != nil {
		var parentFolder Folder
		if err := m.db.Where("id = ? AND user_id = ?", folder.ParentFolderID, folder.UserID).
			First(&parentFolder).Error; err != nil {
			return fmt.Errorf("invalid parent folder")
		}
	}

	return m.db.Create(folder).Error
}

// UpdateFolder updates a folder's details
func (m *FolderModel) UpdateFolder(folder *Folder) error {
	result := m.db.Model(folder).Updates(map[string]interface{}{
		"name":        folder.Name,
		"is_archived": folder.IsArchived,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update folder: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("folder not found")
	}
	return nil
}

// Delete folder and all its contents
func (m *FolderModel) DeleteFolder(folderID, userID uint) error {
	tx := m.db.Begin()

	if err := tx.Model(&Folder{}).
		Where("id = ? OR parent_folder_id = ?", folderID, folderID).
		Update("is_archived", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to archive folder: %w", err)
	}

	if err := tx.Model(&File{}).
		Where("folder_id = ?", folderID).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return tx.Commit().Error
}

// GetFolderPath gets the full path of a folder
func (m *FolderModel) GetFolderPath(folderID uint) ([]Folder, error) {
	var path []Folder
	var currentFolder Folder

	for {
		if err := m.db.First(&currentFolder, folderID).Error; err != nil {
			return nil, fmt.Errorf("failed to get folder path: %w", err)
		}

		path = append([]Folder{currentFolder}, path...) // Prepend to maintain order

		if currentFolder.ParentFolderID == nil {
			break
		}
		folderID = *currentFolder.ParentFolderID
	}

	return path, nil
}

// CreateFolderPath creates the folder hierarchy if it doesn't exist and returns the final folder ID
func (m *FolderModel) CreateFolderPath(userID uint, folderPath string) (*uint, error) {
	if folderPath == "" || folderPath == "/" {
		return nil, nil 
	}

	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var currentParentID *uint
	folders := filepath.SplitList(folderPath)

	for _, folderName := range folders {
		var existingFolder Folder
		err := tx.Where("user_id = ? AND name = ? AND parent_folder_id = ? AND is_archived = ?",
			userID, folderName, currentParentID, false).First(&existingFolder).Error

		if err == gorm.ErrRecordNotFound {
			newFolder := &Folder{
				UserID:         userID,
				Name:           folderName,
				ParentFolderID: currentParentID,
			}

			if err := tx.Create(newFolder).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create folder %s: %w", folderName, err)
			}
			currentParentID = &newFolder.ID
		} else if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("database error while checking folder %s: %w", folderName, err)
		} else {
			currentParentID = &existingFolder.ID
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit folder creation transaction: %w", err)
	}

	return currentParentID, nil
}

// GetFolderByID gets a folder with user verification
func (m *FolderModel) GetFolderByID(folderID, userID uint) (*Folder, error) {
	var folder Folder
	err := m.db.Where("id = ? AND user_id = ? AND is_archived = ?", folderID, userID, false).
		First(&folder).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("folder not found or access denied")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &folder, nil
}
