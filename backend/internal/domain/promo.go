package domain

import (
	"context"
	"errors"
	"time"
)

type PromoCode struct {
	ID             int        `gorm:"primaryKey" json:"id"`
	Code           string     `gorm:"uniqueIndex;not null" json:"code"`
	DiscountType   string     `gorm:"not null" json:"discountType"`
	DiscountValue  float64    `gorm:"type:decimal(10,2);not null" json:"discountValue"`
	MinOrderAmount float64    `gorm:"type:decimal(10,2);default:0" json:"minOrderAmount"`
	MaxUses        *int       `json:"maxUses,omitempty"`
	UsedCount      int        `gorm:"default:0" json:"usedCount"`
	IsActive       bool       `gorm:"default:true" json:"isActive"`
	StartsAt       *time.Time `json:"startsAt,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	Description    *string    `json:"description,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (PromoCode) TableName() string {
	return "promo_codes"
}

type PromoValidationResult struct {
	Valid          bool    `json:"valid"`
	Code           string  `json:"code"`
	DiscountType   string  `json:"discountType"`
	DiscountValue  float64 `json:"discountValue"`
	DiscountAmount float64 `json:"discountAmount"`
	Message        string  `json:"message,omitempty"`
}

var (
	ErrPromoNotFound  = errors.New("promo code not found")
	ErrPromoExpired   = errors.New("promo code has expired")
	ErrPromoUsedUp    = errors.New("promo code usage limit reached")
	ErrPromoInactive  = errors.New("promo code is not active")
	ErrPromoMinAmount = errors.New("order total below minimum amount")
	ErrPromoNotStarted = errors.New("promo code is not yet active")
)

type PromoRepository interface {
	FindByCode(ctx context.Context, code string) (*PromoCode, error)
	FindByID(ctx context.Context, id int) (*PromoCode, error)
	List(ctx context.Context) ([]PromoCode, error)
	Create(ctx context.Context, promo *PromoCode) error
	Update(ctx context.Context, promo *PromoCode) error
	Delete(ctx context.Context, id int) error
	IncrementUsedCount(ctx context.Context, id int) error
}
