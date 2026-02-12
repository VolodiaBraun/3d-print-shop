package domain

import (
	"context"
	"time"
)

type LoyaltySettings struct {
	ID                   int       `gorm:"primaryKey" json:"id"`
	ReferrerBonusPercent float64   `gorm:"type:decimal(5,2);not null" json:"referrerBonusPercent"`
	ReferralWelcomeBonus float64   `gorm:"type:decimal(10,2);not null" json:"referralWelcomeBonus"`
	IsActive             bool      `gorm:"default:true" json:"isActive"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

func (LoyaltySettings) TableName() string {
	return "loyalty_settings"
}

type LoyaltySettingsRepository interface {
	Get(ctx context.Context) (*LoyaltySettings, error)
	Update(ctx context.Context, settings *LoyaltySettings) error
}
