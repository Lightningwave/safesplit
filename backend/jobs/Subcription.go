package jobs

import (
	"time"

	"gorm.io/gorm"
)

type SubscriptionHandler struct {
	db *gorm.DB
}

func NewSubscriptionHandler(db *gorm.DB) *SubscriptionHandler {
	return &SubscriptionHandler{db: db}
}

func (h *SubscriptionHandler) ProcessExpiredSubscriptions() error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		result := tx.Exec(`
            UPDATE users u
            INNER JOIN billing_profiles bp ON u.id = bp.user_id
            SET u.subscription_status = 'free',
                u.role = 'end_user',
                bp.billing_status = 'failed',
                bp.next_billing_date = NULL
            WHERE bp.billing_status = 'cancelled'
            AND bp.next_billing_date < ?
        `, now)

		return result.Error
	})
}

func StartSubscriptionScheduler(handler *SubscriptionHandler) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			if err := handler.ProcessExpiredSubscriptions(); err != nil {
			}
		}
	}()
}
