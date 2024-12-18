package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	ChangedAt    time.Time `json:"changed_at" gorm:"autoCreateTime"`
	User         User      `json:"-" gorm:"foreignKey:UserID"`
}

// TableName overrides the default table name used by GORM.
func (PasswordHistory) TableName() string {
	return "password_history"
}

type PasswordHistoryModel struct {
	db *gorm.DB
}

func NewPasswordHistoryModel(db *gorm.DB) *PasswordHistoryModel {
	return &PasswordHistoryModel{db: db}
}

// AddEntry adds a new password history entry
func (m *PasswordHistoryModel) AddEntry(userID uint, passwordHash string) error {
	entry := &PasswordHistory{
		UserID:       userID,
		PasswordHash: passwordHash,
	}
	return m.db.Create(entry).Error
}

// GetRecentPasswords retrieves the most recent password hashes for a user
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

// CleanupOldEntries removes password history entries older than the specified duration
func (m *PasswordHistoryModel) CleanupOldEntries(userID uint, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	return m.db.Where("user_id = ? AND changed_at < ?", userID, cutoffTime).
		Delete(&PasswordHistory{}).Error
}

// CountUserEntries counts the number of password history entries for a user
func (m *PasswordHistoryModel) CountUserEntries(userID uint) (int64, error) {
	var count int64
	err := m.db.Model(&PasswordHistory{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (m *PasswordHistoryModel) IsPasswordReused(userID uint, newPasswordHash string) (bool, error) {
	recentPasswords, err := m.GetRecentPasswords(userID, 5) // Check last 5 passwords
	if err != nil {
		return false, err
	}

	for _, oldHash := range recentPasswords {
		if oldHash == newPasswordHash {
			return true, nil
		}
	}
	return false, nil
}

func (m *PasswordHistoryModel) ScheduleCleanup(duration time.Duration) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	go func() {
		for range ticker.C {
			// Get all users
			var users []User
			if err := m.db.Find(&users).Error; err != nil {
				continue
			}

			// Clean up old entries for each user
			for _, user := range users {
				_ = m.CleanupOldEntries(user.ID, duration)
			}
		}
	}()
}
