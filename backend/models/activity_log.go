package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id"`
	ActivityType string    `json:"activity_type" gorm:"type:enum('upload','download','delete','share','login','logout','archive','restore','encrypt','decrypt')"`
	FileID       *uint     `json:"file_id,omitempty"`
	FolderID     *uint     `json:"folder_id,omitempty"`
	IPAddress    string    `json:"ip_address"`
	Status       string    `json:"status" gorm:"type:enum('success','failure')"`
	ErrorMessage string    `json:"error_message,omitempty"`
	Details      string    `json:"details,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
}

type ActivityLogModel struct {
	db *gorm.DB
}

func NewActivityLogModel(db *gorm.DB) *ActivityLogModel {
	return &ActivityLogModel{db: db}
}

func (m *ActivityLogModel) GetSystemLogs(filters map[string]interface{}, page, pageSize int) ([]ActivityLog, int64, error) {
	var logs []ActivityLog
	var total int64

	query := m.db.Model(&ActivityLog{}).Preload("User")

	// Apply filters
	if timestamp, ok := filters["timestamp"].(string); ok {
		query = query.Where("DATE(created_at) = ?", timestamp)
	}
	if category, ok := filters["activity_type"].(string); ok {
		query = query.Where("activity_type = ?", category)
	}
	if source, ok := filters["source"].(string); ok {
		query = query.Where("activity_type LIKE ?", "%"+source+"%")
	}
	if userID, ok := filters["user_id"].(uint); ok {
		query = query.Where("user_id = ?", userID)
	}

	// Get total count
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

func (m *ActivityLogModel) LogActivity(log *ActivityLog) error {
	return m.db.Create(log).Error
}
