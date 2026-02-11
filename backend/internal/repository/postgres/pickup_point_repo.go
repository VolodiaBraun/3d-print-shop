package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type PickupPointRepo struct {
	db *gorm.DB
}

func NewPickupPointRepo(db *gorm.DB) *PickupPointRepo {
	return &PickupPointRepo{db: db}
}

func (r *PickupPointRepo) FindByID(ctx context.Context, id int) (*domain.PickupPoint, error) {
	var point domain.PickupPoint
	err := r.db.WithContext(ctx).First(&point, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrPickupPointNotFound
	}
	return &point, err
}

func (r *PickupPointRepo) FindByCity(ctx context.Context, city string) ([]domain.PickupPoint, error) {
	city = strings.TrimSpace(city)
	var points []domain.PickupPoint
	err := r.db.WithContext(ctx).
		Where("is_active = true AND city ILIKE ?", city).
		Order("name ASC").
		Find(&points).Error
	return points, err
}

func (r *PickupPointRepo) List(ctx context.Context) ([]domain.PickupPoint, error) {
	var points []domain.PickupPoint
	err := r.db.WithContext(ctx).Order("city ASC, name ASC").Find(&points).Error
	return points, err
}

func (r *PickupPointRepo) Create(ctx context.Context, point *domain.PickupPoint) error {
	return r.db.WithContext(ctx).Create(point).Error
}

func (r *PickupPointRepo) Update(ctx context.Context, point *domain.PickupPoint) error {
	return r.db.WithContext(ctx).Save(point).Error
}

func (r *PickupPointRepo) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&domain.PickupPoint{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPickupPointNotFound
	}
	return nil
}
