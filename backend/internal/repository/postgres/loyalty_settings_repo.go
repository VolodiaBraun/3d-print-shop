package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type LoyaltySettingsRepo struct {
	db *gorm.DB
}

func NewLoyaltySettingsRepo(db *gorm.DB) *LoyaltySettingsRepo {
	return &LoyaltySettingsRepo{db: db}
}

func (r *LoyaltySettingsRepo) Get(ctx context.Context) (*domain.LoyaltySettings, error) {
	var settings domain.LoyaltySettings
	err := r.db.WithContext(ctx).First(&settings).Error
	return &settings, err
}

func (r *LoyaltySettingsRepo) Update(ctx context.Context, settings *domain.LoyaltySettings) error {
	return r.db.WithContext(ctx).Save(settings).Error
}
