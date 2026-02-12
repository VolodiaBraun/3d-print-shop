package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// User represents a user in the system.
type User struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Email        *string   `gorm:"uniqueIndex:idx_users_email,where:email IS NOT NULL" json:"email,omitempty"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	TelegramID   *int64    `gorm:"uniqueIndex" json:"telegramId,omitempty"`
	Phone        *string   `json:"phone,omitempty"`
	FirstName    *string   `json:"firstName,omitempty"`
	LastName     *string   `json:"lastName,omitempty"`
	Username     *string   `json:"username,omitempty"`
	Role             string    `gorm:"default:customer" json:"role"`
	IsActive         bool      `gorm:"default:true" json:"isActive"`
	ReferralCode     *string   `gorm:"uniqueIndex" json:"referralCode,omitempty"`
	ReferredByUserID *int      `json:"referredByUserID,omitempty"`
	BonusBalance     float64   `gorm:"default:0" json:"bonusBalance"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}
