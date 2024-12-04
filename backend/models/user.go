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

type UserModel struct {
	db *gorm.DB
}

func NewUserModel(db *gorm.DB) *UserModel {
	return &UserModel{db: db}
}

// AuthenticateUser checks the provided email and password against the database
func (m *UserModel) Authenticate(email, password string) (*User, error) {
	var user User
	if err := m.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}

// Create creates a new user in the database
func (m *UserModel) Create(user *User) (*User, error) {
	if err := m.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// FindByEmail retrieves a user by their email
func (u *User) FindByEmail(db *gorm.DB, email string) error {
	return db.Where("email = ?", email).First(u).Error
}

// FindByID retrieves a user by their ID
func (m *UserModel) FindByID(id uint) (*User, error) {
	var user User
	if err := m.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// IsEndUser checks if the user has the "end_user" role
func (u *User) IsEndUser() bool {
	return u.Role == RoleEndUser
}

// IsPremiumUser checks if the user has the "premium_user" role
func (u *User) IsPremiumUser() bool {
	return u.Role == RolePremiumUser
}

// IsSysAdmin checks if the user has the "sys_admin" role
func (u *User) IsSysAdmin() bool {
	return u.Role == RoleSysAdmin
}

// IsSuperAdmin checks if the user has the "super_admin" role
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// UpgradeToPremium updates the user's role to "premium_user"
func (u *User) UpgradeToPremium(db *gorm.DB) error {
	u.Role = RolePremiumUser
	return db.Save(u).Error
}
