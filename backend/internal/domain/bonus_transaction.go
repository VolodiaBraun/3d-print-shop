package domain

import (
	"context"
	"time"
)

type BonusTransaction struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	UserID      int       `gorm:"not null" json:"userId"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Type        string    `gorm:"not null" json:"type"`
	ReferenceID *int      `json:"referenceId,omitempty"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (BonusTransaction) TableName() string {
	return "bonus_transactions"
}

type BonusTransactionRepository interface {
	Create(ctx context.Context, tx *BonusTransaction) error
	CreateWithDB(ctx context.Context, db interface{}, tx *BonusTransaction) error
	ListByUserID(ctx context.Context, userID int) ([]BonusTransaction, error)
}
