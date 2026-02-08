package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type PromoRepo struct {
	db *gorm.DB
}

func NewPromoRepo(db *gorm.DB) *PromoRepo {
	return &PromoRepo{db: db}
}

func (r *PromoRepo) FindByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	var promo domain.PromoCode
	err := r.db.WithContext(ctx).
		Where("UPPER(code) = UPPER(?)", strings.TrimSpace(code)).
		First(&promo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrPromoNotFound
	}
	return &promo, err
}

func (r *PromoRepo) FindByID(ctx context.Context, id int) (*domain.PromoCode, error) {
	var promo domain.PromoCode
	err := r.db.WithContext(ctx).First(&promo, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrPromoNotFound
	}
	return &promo, err
}

func (r *PromoRepo) List(ctx context.Context) ([]domain.PromoCode, error) {
	var promos []domain.PromoCode
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&promos).Error
	return promos, err
}

func (r *PromoRepo) Create(ctx context.Context, promo *domain.PromoCode) error {
	promo.Code = strings.ToUpper(strings.TrimSpace(promo.Code))
	return r.db.WithContext(ctx).Create(promo).Error
}

func (r *PromoRepo) Update(ctx context.Context, promo *domain.PromoCode) error {
	promo.Code = strings.ToUpper(strings.TrimSpace(promo.Code))
	return r.db.WithContext(ctx).Save(promo).Error
}

func (r *PromoRepo) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&domain.PromoCode{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPromoNotFound
	}
	return nil
}

func (r *PromoRepo) IncrementUsedCount(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).
		Model(&domain.PromoCode{}).
		Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}
