package jobs

import (
    "log"
    "time"
    "gorm.io/gorm"
)

// Constants for job scheduling and thresholds
const (
    InactivityThreshold = 90 * 24 * time.Hour  // 90 days
    DeletionThreshold  = 180 * 24 * time.Hour  // 180 days
    
    AccountProcessingInterval = 1 * time.Hour
    SubscriptionInterval     = 24 * time.Hour
)

type User struct {
    ID                 uint
    IsActive          bool
    LastLogin         *time.Time
    UpdatedAt         time.Time
    Role              string
    AccountLockedUntil *time.Time
    SubscriptionStatus string
    StorageQuota      int64
}

type JobManager struct {
    db               *gorm.DB
    accountManager   *AccountManager
    subHandler      *SubscriptionHandler
}

func NewJobManager(db *gorm.DB) *JobManager {
    return &JobManager{
        db:              db,
        accountManager:  NewAccountManager(db),
        subHandler:     NewSubscriptionHandler(db),
    }
}

func (m *JobManager) StartAllJobs() {
    m.StartAccountManagementJob()
    m.StartSubscriptionJob()
    log.Println("All scheduled jobs started")
}

func (m *JobManager) StartAccountManagementJob() {
    ticker := time.NewTicker(AccountProcessingInterval)
    go func() {
        for range ticker.C {
            if err := m.accountManager.ProcessAccounts(); err != nil {
                log.Printf("Error in account management job: %v", err)
            }
        }
    }()
    log.Println("Account management job started")
}

func (m *JobManager) StartSubscriptionJob() {
    ticker := time.NewTicker(SubscriptionInterval)
    go func() {
        for range ticker.C {
            if err := m.subHandler.ProcessExpiredSubscriptions(); err != nil {
                log.Printf("Error in subscription processing job: %v", err)
            }
        }
    }()
    log.Println("Subscription processing job started")
}

type AccountManager struct {
    db *gorm.DB
}

func NewAccountManager(db *gorm.DB) *AccountManager {
    return &AccountManager{db: db}
}

func (m *AccountManager) ProcessAccounts() error {
    if err := m.unlockExpiredAccounts(); err != nil {
        log.Printf("Error unlocking accounts: %v", err)
    }

    if err := m.deactivateInactiveAccounts(); err != nil {
        log.Printf("Error deactivating accounts: %v", err)
    }

    if err := m.deleteInactiveAccounts(); err != nil {
        log.Printf("Error deleting accounts: %v", err)
    }

    return nil
}

func (m *AccountManager) unlockExpiredAccounts() error {
    result := m.db.Exec(`
        UPDATE users 
        SET account_locked_until = NULL,
            failed_login_attempts = 0
        WHERE account_locked_until < ? 
        AND account_locked_until IS NOT NULL
        AND is_active = true`,
        time.Now(),
    )

    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected > 0 {
        log.Printf("Unlocked %d accounts", result.RowsAffected)
    }
    return nil
}

func (m *AccountManager) deactivateInactiveAccounts() error {
    result := m.db.Exec(`
        UPDATE users 
        SET is_active = false,
            subscription_status = CASE 
                WHEN subscription_status = 'premium' THEN 'cancelled'
                ELSE subscription_status 
            END,
            storage_quota = CASE 
                WHEN subscription_status = 'premium' THEN 5368709120  -- 5GB
                ELSE storage_quota 
            END
        WHERE last_login < ?
        AND is_active = true
        AND role NOT IN ('sys_admin', 'super_admin')`,
        time.Now().Add(-InactivityThreshold),
    )

    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected > 0 {
        log.Printf("Deactivated %d inactive accounts", result.RowsAffected)
    }
    return nil
}

func (m *AccountManager) deleteInactiveAccounts() error {
    return m.db.Transaction(func(tx *gorm.DB) error {
        // First, get the IDs of accounts to be deleted
        var userIDs []uint
        if err := tx.Model(&User{}).
            Where("is_active = ? AND updated_at < ? AND role NOT IN ?",
                false,
                time.Now().Add(-DeletionThreshold),
                []string{"sys_admin", "super_admin"}).
            Pluck("id", &userIDs).Error; err != nil {
            return err
        }

        if len(userIDs) == 0 {
            return nil
        }

        deleteQueries := []string{
            "DELETE FROM password_history WHERE user_id IN (?)",
            "DELETE FROM activity_logs WHERE user_id IN (?)",
            "DELETE FROM billing_profiles WHERE user_id IN (?)",
            "DELETE FROM key_fragments WHERE user_id IN (?)",
            "DELETE FROM user_files WHERE user_id IN (?)",
            "DELETE FROM users WHERE id IN (?)",
        }

        for _, query := range deleteQueries {
            if err := tx.Exec(query, userIDs).Error; err != nil {
                return err
            }
        }

        log.Printf("Permanently deleted %d inactive accounts", len(userIDs))
        return nil
    })
}

type SubscriptionHandler struct {
    db *gorm.DB
}

func NewSubscriptionHandler(db *gorm.DB) *SubscriptionHandler {
    return &SubscriptionHandler{db: db}
}

func (h *SubscriptionHandler) ProcessExpiredSubscriptions() error {
    return h.db.Transaction(func(tx *gorm.DB) error {
        result := tx.Exec(`
            UPDATE users u
            INNER JOIN billing_profiles bp ON u.id = bp.user_id
            SET u.subscription_status = 'free',
                u.role = 'end_user',
                u.storage_quota = 5368709120,  -- 5GB
                bp.billing_status = 'failed',
                bp.next_billing_date = NULL
            WHERE bp.billing_status = 'cancelled'
            AND bp.next_billing_date < ?
            AND u.subscription_status = 'premium'`,
            time.Now(),
        )

        if result.Error != nil {
            return result.Error
        }

        if result.RowsAffected > 0 {
            log.Printf("Processed %d expired subscriptions", result.RowsAffected)
        }
        return nil
    })
}