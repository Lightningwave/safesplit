package models

import (
	"time"

	"gorm.io/gorm"
)

type FeedbackType string
type FeedbackStatus string

const (
	FeedbackTypeFeedback           FeedbackType = "feedback"
	FeedbackTypeSuspiciousActivity FeedbackType = "suspicious_activity"

	FeedbackStatusPending   FeedbackStatus = "pending"
	FeedbackStatusInReview  FeedbackStatus = "in_review"
	FeedbackStatusResolved  FeedbackStatus = "resolved"
)

// Feedback represents the feedback table in the database
type Feedback struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint          `json:"user_id"`
	Type      FeedbackType  `json:"type" gorm:"type:enum('feedback','suspicious_activity')"`
	Subject   string        `json:"subject" gorm:"size:255;not null"`
	Message   string        `json:"message" gorm:"type:text;not null"`
	Status    FeedbackStatus `json:"status" gorm:"type:enum('pending','in_review','resolved');default:pending"`
	CreatedAt time.Time     `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP;ON UPDATE CURRENT_TIMESTAMP"`
	User      User          `json:"user" gorm:"foreignKey:UserID"`
}

// FeedbackModel handles database operations for feedback
type FeedbackModel struct {
	db *gorm.DB
}

// NewFeedbackModel creates a new FeedbackModel instance
func NewFeedbackModel(db *gorm.DB) *FeedbackModel {
	return &FeedbackModel{db: db}
}

// Create adds a new feedback entry
func (m *FeedbackModel) Create(feedback *Feedback) error {
	return m.db.Create(feedback).Error
}

// GetByID retrieves a feedback entry by its ID
func (m *FeedbackModel) GetByID(id uint) (*Feedback, error) {
	var feedback Feedback
	if err := m.db.Preload("User").First(&feedback, id).Error; err != nil {
		return nil, err
	}
	return &feedback, nil
}

// GetAllByUser retrieves all feedback entries for a specific user
func (m *FeedbackModel) GetAllByUser(userID uint) ([]Feedback, error) {
	var feedbacks []Feedback
	if err := m.db.Where("user_id = ?", userID).Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}

// GetAll retrieves all feedback entries with optional filters
func (m *FeedbackModel) GetAll(filters map[string]interface{}, page, pageSize int) ([]Feedback, int64, error) {
	var feedbacks []Feedback
	var total int64

	query := m.db.Model(&Feedback{}).Preload("User")

	// Apply filters
	if feedbackType, ok := filters["type"].(string); ok {
		query = query.Where("type = ?", feedbackType)
	}
	if status, ok := filters["status"].(string); ok {
		query = query.Where("status = ?", status)
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
		Find(&feedbacks).Error

	return feedbacks, total, err
}

// UpdateStatus updates the status of a feedback entry
func (m *FeedbackModel) UpdateStatus(id uint, status FeedbackStatus) error {
	return m.db.Model(&Feedback{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Delete removes a feedback entry
func (m *FeedbackModel) Delete(id uint) error {
	return m.db.Delete(&Feedback{}, id).Error
}

// GetByStatus retrieves all feedback entries with a specific status
func (m *FeedbackModel) GetByStatus(status FeedbackStatus) ([]Feedback, error) {
	var feedbacks []Feedback
	if err := m.db.Where("status = ?", status).Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}

// GetByType retrieves all feedback entries of a specific type
func (m *FeedbackModel) GetByType(feedbackType FeedbackType) ([]Feedback, error) {
	var feedbacks []Feedback
	if err := m.db.Where("type = ?", feedbackType).Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}

// GetPendingCount returns the count of pending feedback entries
func (m *FeedbackModel) GetPendingCount() (int64, error) {
	var count int64
	err := m.db.Model(&Feedback{}).
		Where("status = ?", FeedbackStatusPending).
		Count(&count).Error
	return count, err
}

// GetDateRangeCount returns the count of feedback entries within a date range
func (m *FeedbackModel) GetDateRangeCount(startDate, endDate time.Time) (int64, error) {
	var count int64
	err := m.db.Model(&Feedback{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&count).Error
	return count, err
}