package models

import (
	"time"
    "fmt"
	"gorm.io/gorm"
)

type PasswordHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	ChangedAt    time.Time `json:"changed_at" gorm:"autoCreateTime"`
	User         User      `json:"-" gorm:"foreignKey:UserID"`
}

func (PasswordHistory) TableName() string {
	return "password_history"
}

type PasswordHistoryModel struct {
	db *gorm.DB
}

func NewPasswordHistoryModel(db *gorm.DB) *PasswordHistoryModel {
	return &PasswordHistoryModel{db: db}
}

func (m *PasswordHistoryModel) AddEntry(userID uint, passwordHash string) error {
	entry := &PasswordHistory{
		UserID:       userID,
		PasswordHash: passwordHash,
	}
	return m.db.Create(entry).Error
}

func (m *PasswordHistoryModel) GetRecentPasswords(userID uint, limit int) ([]string, error) {
	var entries []PasswordHistory
	if err := m.db.Where("user_id = ?", userID).
		Order("changed_at DESC").
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	hashes := make([]string, len(entries))
	for i, entry := range entries {
		hashes[i] = entry.PasswordHash
	}
	return hashes, nil
}
func (m *PasswordHistoryModel) IsPasswordReused(userID uint, newPasswordHash string) (bool, error) {
    recentPasswords, err := m.GetRecentPasswords(userID, 5)
    if err != nil {
        return false, fmt.Errorf("failed to get recent passwords: %w", err)
    }

    for _, oldHash := range recentPasswords {
        if oldHash == newPasswordHash {
            return true, nil
        }
    }

    return false, nil
}



