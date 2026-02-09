package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepo) FindByID(ctx context.Context, id int) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images", "is_main = true").
		First(&order, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrOrderNotFound
	}
	return &order, err
}

func (r *OrderRepo) FindByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images", "is_main = true").
		Where("order_number = ?", orderNumber).
		First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrOrderNotFound
	}
	return &order, err
}

func (r *OrderRepo) List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Order{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	listQuery := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images", "is_main = true")

	if filter.Status != "" {
		listQuery = listQuery.Where("status = ?", filter.Status)
	}

	var orders []domain.Order
	err := listQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepo) ListByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepo) UpdateTracking(ctx context.Context, id int, trackingNumber string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"tracking_number": trackingNumber,
			"updated_at":      time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepo) NextOrderNumber(ctx context.Context) (string, error) {
	today := time.Now().Format("20060102")

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("order_number LIKE ?", fmt.Sprintf("ORD-%s-%%", today)).
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ORD-%s-%04d", today, count+1), nil
}
