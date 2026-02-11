package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type DeliveryZoneRepo struct {
	db *gorm.DB
}

func NewDeliveryZoneRepo(db *gorm.DB) *DeliveryZoneRepo {
	return &DeliveryZoneRepo{db: db}
}

func (r *DeliveryZoneRepo) FindByID(ctx context.Context, id int) (*domain.DeliveryZone, error) {
	var zone domain.DeliveryZone
	err := r.db.WithContext(ctx).First(&zone, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrDeliveryZoneNotFound
	}
	return &zone, err
}

func (r *DeliveryZoneRepo) FindByCity(ctx context.Context, city string) (*domain.DeliveryZone, error) {
	city = strings.TrimSpace(city)
	var zone domain.DeliveryZone
	err := r.db.WithContext(ctx).
		Where("is_active = true AND city ILIKE ?", city).
		First(&zone).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrDeliveryZoneNotFound
	}
	return &zone, err
}

func (r *DeliveryZoneRepo) List(ctx context.Context) ([]domain.DeliveryZone, error) {
	var zones []domain.DeliveryZone
	err := r.db.WithContext(ctx).Order("name ASC").Find(&zones).Error
	return zones, err
}

func (r *DeliveryZoneRepo) Create(ctx context.Context, zone *domain.DeliveryZone) error {
	return r.db.WithContext(ctx).Create(zone).Error
}

func (r *DeliveryZoneRepo) Update(ctx context.Context, zone *domain.DeliveryZone) error {
	return r.db.WithContext(ctx).Save(zone).Error
}

func (r *DeliveryZoneRepo) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&domain.DeliveryZone{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrDeliveryZoneNotFound
	}
	return nil
}
