package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Define user roles as constants
const (
	RoleEndUser     = "end_user"
	RolePremiumUser = "premium_user"
	RoleSysAdmin    = "sys_admin"
	RoleSuperAdmin  = "super_admin"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"password" gorm:"not null"`
	Role      string    `json:"role" gorm:"type:enum('end_user','premium_user','sys_admin','super_admin');default:'end_user'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Authenticate user
func AuthenticateUser(db *gorm.DB, email, password string) (*User, error) {
	var user User

	// Find user by email
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

// Database operations
func (u *User) Create(db *gorm.DB) error {
	// Set default role if not specified
	if u.Role == "" {
		u.Role = RoleEndUser
	}

	// Hash password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	return db.Create(u).Error
}

func (u *User) FindByEmail(db *gorm.DB, email string) error {
	return db.Where("email = ?", email).First(u).Error
}

func (u *User) FindByID(db *gorm.DB, id uint) error {
	return db.First(u, id).Error
}

// Helper methods for role checks
func (u *User) IsEndUser() bool {
	return u.Role == RoleEndUser
}

func (u *User) IsPremiumUser() bool {
	return u.Role == RolePremiumUser
}

func (u *User) IsSysAdmin() bool {
	return u.Role == RoleSysAdmin
}

func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// Method to upgrade to premium
func (u *User) UpgradeToPremium(db *gorm.DB) error {
	u.Role = RolePremiumUser
	return db.Save(u).Error
}
