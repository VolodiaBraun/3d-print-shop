package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type BonusTransactionRepo struct {
	db *gorm.DB
}

func NewBonusTransactionRepo(db *gorm.DB) *BonusTransactionRepo {
	return &BonusTransactionRepo{db: db}
}

func (r *BonusTransactionRepo) Create(ctx context.Context, tx *domain.BonusTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *BonusTransactionRepo) CreateWithDB(ctx context.Context, db interface{}, tx *domain.BonusTransaction) error {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return r.Create(ctx, tx)
	}
	return gormDB.WithContext(ctx).Create(tx).Error
}

func (r *BonusTransactionRepo) ListByUserID(ctx context.Context, userID int) ([]domain.BonusTransaction, error) {
	var transactions []domain.BonusTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}
