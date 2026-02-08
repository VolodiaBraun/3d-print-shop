package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type CartRepo struct {
	db *gorm.DB
}

func NewCartRepo(db *gorm.DB) *CartRepo {
	return &CartRepo{db: db}
}

func (r *CartRepo) GetByUserID(ctx context.Context, userID int) ([]domain.CartItem, error) {
	var items []domain.CartItem
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Product").
		Preload("Product.Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = true").Order("display_order ASC").Limit(1)
		}).
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *CartRepo) FindItem(ctx context.Context, userID int, productID int) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCartItemNotFound
	}
	return &item, err
}

func (r *CartRepo) FindItemByID(ctx context.Context, id int, userID int) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCartItemNotFound
	}
	return &item, err
}

func (r *CartRepo) AddItem(ctx context.Context, item *domain.CartItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *CartRepo) UpdateQuantity(ctx context.Context, id int, userID int, quantity int) error {
	result := r.db.WithContext(ctx).
		Model(&domain.CartItem{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("quantity", quantity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrCartItemNotFound
	}
	return nil
}

func (r *CartRepo) RemoveItem(ctx context.Context, id int, userID int) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.CartItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrCartItemNotFound
	}
	return nil
}

func (r *CartRepo) Clear(ctx context.Context, userID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.CartItem{}).Error
}
