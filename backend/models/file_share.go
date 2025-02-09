package models

import (
   "crypto/rand"
   "encoding/base64"
   "fmt"
   "time"
   "golang.org/x/crypto/bcrypt"
   "gorm.io/gorm"
)

type ShareType string

const (
   NormalShare    ShareType = "normal"
   RecipientShare ShareType = "recipient"
)

type FileShare struct {
   ID                   uint       `json:"id" gorm:"primaryKey"`
   FileID               uint       `json:"file_id"`
   SharedBy             uint       `json:"shared_by"`
   ShareLink            string     `json:"share_link" gorm:"unique"`
   PasswordHash         string     `json:"-"`
   PasswordSalt         string     `json:"-"`
   EncryptedKeyFragment []byte     `json:"-" gorm:"type:mediumblob"`
   FragmentIndex        int        `json:"-" gorm:"not null"`
   ExpiresAt            *time.Time `json:"expires_at"`
   MaxDownloads         *int       `json:"max_downloads"`
   DownloadCount        int        `json:"download_count" gorm:"default:0"`
   IsActive             bool       `json:"is_active" gorm:"default:true"`
   CreatedAt            time.Time  `json:"created_at"`
   File                 File       `json:"file" gorm:"foreignKey:FileID"`
   ShareType            ShareType  `json:"share_type" gorm:"type:varchar(20);default:'normal'"`
   Email                string     `json:"email,omitempty"`
}

type FileShareModel struct {
   db *gorm.DB
}

func NewFileShareModel(db *gorm.DB) *FileShareModel {
   return &FileShareModel{db: db}
}

func generateShareLink() (string, error) {
   bytes := make([]byte, 32)
   if _, err := rand.Read(bytes); err != nil {
       return "", err
   }
   return base64.URLEncoding.EncodeToString(bytes), nil
}

func (m *FileShareModel) CreateFileShare(share *FileShare, password string) error {
   if share.ShareType == RecipientShare && share.Email == "" {
       return fmt.Errorf("email required for recipient share")
   }

   salt := make([]byte, 16)
   if _, err := rand.Read(salt); err != nil {
       return fmt.Errorf("failed to generate salt: %w", err)
   }
   share.PasswordSalt = base64.StdEncoding.EncodeToString(salt)

   hashedPassword, err := bcrypt.GenerateFromPassword(
       []byte(password+share.PasswordSalt),
       bcrypt.DefaultCost,
   )
   if err != nil {
       return fmt.Errorf("failed to hash password: %w", err)
   }
   share.PasswordHash = string(hashedPassword)

   shareLink, err := generateShareLink()
   if err != nil {
       return fmt.Errorf("failed to generate share link: %w", err)
   }
   share.ShareLink = shareLink

   tx := m.db.Begin()
   if tx.Error != nil {
       return fmt.Errorf("failed to start transaction: %w", tx.Error)
   }

   if err := tx.Create(share).Error; err != nil {
       tx.Rollback()
       return fmt.Errorf("failed to create share record: %w", err)
   }

   if err := tx.Model(&File{}).Where("id = ?", share.FileID).Update("is_shared", true).Error; err != nil {
       tx.Rollback()
       return fmt.Errorf("failed to update file status: %w", err)
   }

   return tx.Commit().Error
}

func (m *FileShareModel) ValidateShare(shareLink string, password string) (*FileShare, error) {
   var share FileShare
   if err := m.db.Where("share_link = ? AND is_active = ? AND share_type = ?", 
       shareLink, true, NormalShare).Preload("File").First(&share).Error; err != nil {
       return nil, fmt.Errorf("share not found or inactive")
   }

   if err := bcrypt.CompareHashAndPassword(
       []byte(share.PasswordHash),
       []byte(password+share.PasswordSalt),
   ); err != nil {
       return nil, fmt.Errorf("invalid password")
   }

   return &share, nil
}

func (m *FileShareModel) ValidateRecipientShare(shareLink string, password string) (*FileShare, error) {
	var share FileShare
	if err := m.db.Where("share_link = ? AND is_active = ? AND share_type = ?", 
		shareLink, true, RecipientShare).Preload("File").First(&share).Error; err != nil {
		return nil, fmt.Errorf("share not found or invalid")
	}
 
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		share.IsActive = false
		m.db.Save(&share)
		return nil, fmt.Errorf("share has expired")
	}
 
	if share.MaxDownloads != nil && share.DownloadCount >= *share.MaxDownloads {
		share.IsActive = false
		m.db.Save(&share)
		return nil, fmt.Errorf("download limit exceeded")
	}
 
	if err := bcrypt.CompareHashAndPassword(
		[]byte(share.PasswordHash),
		[]byte(password+share.PasswordSalt),
	); err != nil {
		return nil, fmt.Errorf("invalid password")
	}
 
	return &share, nil
 }
 func (m *FileShareModel) ValidatePassword(shareLink string, password string) error {
    var share FileShare
    if err := m.db.Where("share_link = ?", shareLink).First(&share).Error; err != nil {
        return fmt.Errorf("share not found")
    }

    if err := bcrypt.CompareHashAndPassword(
        []byte(share.PasswordHash),
        []byte(password+share.PasswordSalt),
    ); err != nil {
        return fmt.Errorf("invalid password")
    }

    return nil
}

func (m *FileShareModel) IncrementDownloadCount(shareID uint) error {
   tx := m.db.Begin()
   if tx.Error != nil {
       return fmt.Errorf("failed to start transaction: %w", tx.Error)
   }
   defer tx.Rollback()

   result := tx.Model(&FileShare{}).
       Where("id = ?", shareID).
       Update("download_count", gorm.Expr("download_count + ?", 1))

   if result.Error != nil {
       return fmt.Errorf("failed to increment download count: %w", result.Error)
   }

   if result.RowsAffected == 0 {
       return fmt.Errorf("no share found with ID %d", shareID)
   }

   return tx.Commit().Error
}
func (m *FileShareModel) GetShareByLink(shareLink string) (*FileShare, error) {
    var share FileShare
    err := m.db.Where("share_link = ? AND is_active = ?", shareLink, true).
        Preload("File").First(&share).Error
    if err != nil {
        return nil, fmt.Errorf("share not found or inactive")
    }
    return &share, nil
}