package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type FeedbackType string
type FeedbackStatus string

const (
	FeedbackTypeFeedback           FeedbackType = "feedback"
	FeedbackTypeSuspiciousActivity FeedbackType = "suspicious_activity"

	FeedbackStatusPending  FeedbackStatus = "pending"
	FeedbackStatusInReview FeedbackStatus = "in_review"
	FeedbackStatusResolved FeedbackStatus = "resolved"
)

type Feedback struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id"`
	Type      FeedbackType   `json:"type" gorm:"type:enum('feedback','suspicious_activity')"`
	Subject   string         `json:"subject" gorm:"size:255;not null"`
	Message   string         `json:"message" gorm:"type:text;not null"`
	Details   string         `json:"details" gorm:"type:text"`
	Status    FeedbackStatus `json:"status" gorm:"type:enum('pending','in_review','resolved');default:pending"`
	CreatedAt time.Time      `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP;ON UPDATE CURRENT_TIMESTAMP"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
}

type FeedbackModel struct {
	db *gorm.DB
}

func NewFeedbackModel(db *gorm.DB) *FeedbackModel {
	return &FeedbackModel{db: db}
}

func (m *FeedbackModel) Create(feedback *Feedback) error {
	return m.db.Create(feedback).Error
}

func (m *FeedbackModel) GetByID(id uint) (*Feedback, error) {
	var feedback Feedback
	if err := m.db.Preload("User").First(&feedback, id).Error; err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (m *FeedbackModel) GetAllByUser(userID uint) ([]Feedback, error) {
	var feedbacks []Feedback
	if err := m.db.Where("user_id = ?", userID).Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}

func (m *FeedbackModel) GetAllByUserAndType(userID uint, feedbackType FeedbackType) ([]Feedback, error) {
	var feedbacks []Feedback
	err := m.db.Where("user_id = ? AND type = ?", userID, feedbackType).
		Order("created_at DESC").
		Find(&feedbacks).Error
	if err != nil {
		return nil, err
	}
	return feedbacks, nil
}

func (m *FeedbackModel) GetAll(filters map[string]interface{}, page, pageSize int) ([]Feedback, int64, error) {
	var feedbacks []Feedback
	var total int64

	query := m.db.Model(&Feedback{}).Preload("User")

	if feedbackType, ok := filters["type"].(FeedbackType); ok {
		query = query.Where("type = ?", feedbackType)
	}
	if status, ok := filters["status"].(string); ok {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filters["user_id"].(uint); ok {
		query = query.Where("user_id = ?", userID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&feedbacks).Error

	return feedbacks, total, err
}

func (m *FeedbackModel) UpdateStatus(id uint, status FeedbackStatus) error {
	return m.db.Model(&Feedback{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (m *FeedbackModel) Delete(id uint) error {
	return m.db.Delete(&Feedback{}, id).Error
}

func (m *FeedbackModel) GetByStatus(status FeedbackStatus) ([]Feedback, error) {
	var feedbacks []Feedback
	if err := m.db.Where("status = ?", status).Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}

func (m *FeedbackModel) GetByType(feedbackType FeedbackType) ([]Feedback, error) {
	var feedbacks []Feedback
	if err := m.db.Where("type = ?", feedbackType).Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}

func (m *FeedbackModel) GetPendingCount() (int64, error) {
	var count int64
	err := m.db.Model(&Feedback{}).
		Where("status = ? AND type = ?", FeedbackStatusPending, FeedbackTypeFeedback).
		Count(&count).Error
	return count, err
}

func (m *FeedbackModel) GetDateRangeCount(startDate, endDate time.Time) (int64, error) {
	var count int64
	err := m.db.Model(&Feedback{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&count).Error
	return count, err
}

func (m *FeedbackModel) UpdateStatusWithComment(id uint, status FeedbackStatus, comment string) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		feedback, err := m.GetByID(id)
		if err != nil {
			return fmt.Errorf("feedback not found: %w", err)
		}

		feedback.Status = status

		timestamp := time.Now().Format(time.RFC3339)
		newDetails := fmt.Sprintf("%s\n[%s] Status changed to %s: %s",
			feedback.Details,
			timestamp,
			status,
			comment,
		)
		feedback.Details = newDetails

		if err := tx.Save(feedback).Error; err != nil {
			return fmt.Errorf("failed to update feedback: %w", err)
		}

		return nil
	})
}
