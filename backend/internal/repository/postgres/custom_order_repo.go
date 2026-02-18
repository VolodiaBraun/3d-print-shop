package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type CustomOrderRepo struct {
	db *gorm.DB
}

func NewCustomOrderRepo(db *gorm.DB) *CustomOrderRepo {
	return &CustomOrderRepo{db: db}
}

func (r *CustomOrderRepo) Create(ctx context.Context, details *domain.CustomOrderDetails) error {
	return r.db.WithContext(ctx).Create(details).Error
}

func (r *CustomOrderRepo) FindByOrderID(ctx context.Context, orderID int) (*domain.CustomOrderDetails, error) {
	var details domain.CustomOrderDetails
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&details).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCustomOrderNotFound
	}
	return &details, err
}

func (r *CustomOrderRepo) Update(ctx context.Context, details *domain.CustomOrderDetails) error {
	return r.db.WithContext(ctx).Save(details).Error
}
