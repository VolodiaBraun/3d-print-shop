package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type ReviewRepo struct {
	db *gorm.DB
}

func NewReviewRepo(db *gorm.DB) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (r *ReviewRepo) Create(ctx context.Context, review *domain.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *ReviewRepo) FindByID(ctx context.Context, id int) (*domain.Review, error) {
	var review domain.Review
	err := r.db.WithContext(ctx).Preload("User").Preload("Product").First(&review, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrReviewNotFound
	}
	return &review, err
}

func (r *ReviewRepo) FindByProductID(ctx context.Context, productID int) ([]domain.Review, error) {
	var reviews []domain.Review
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND status = ?", productID, "approved").
		Preload("User").
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *ReviewRepo) FindByUserAndProduct(ctx context.Context, userID, productID int) (*domain.Review, error) {
	var review domain.Review
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&review).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrReviewNotFound
	}
	return &review, err
}

func (r *ReviewRepo) List(ctx context.Context, filter domain.ReviewFilter) ([]domain.Review, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Review{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ProductID != 0 {
		query = query.Where("product_id = ?", filter.ProductID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	var reviews []domain.Review
	err := query.
		Preload("User").
		Preload("Product").
		Order("created_at DESC").
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Find(&reviews).Error

	return reviews, total, err
}

func (r *ReviewRepo) ListByUserID(ctx context.Context, userID int) ([]domain.Review, error) {
	var reviews []domain.Review
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Product").
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *ReviewRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	result := r.db.WithContext(ctx).Model(&domain.Review{}).Where("id = ?", id).Update("status", status)
	if result.RowsAffected == 0 {
		return domain.ErrReviewNotFound
	}
	return result.Error
}

func (r *ReviewRepo) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&domain.Review{}, id)
	if result.RowsAffected == 0 {
		return domain.ErrReviewNotFound
	}
	return result.Error
}

func (r *ReviewRepo) GetProductRatingStats(ctx context.Context, productID int) (float64, int, error) {
	var result struct {
		AvgRating float64
		Count     int
	}
	err := r.db.WithContext(ctx).
		Model(&domain.Review{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("product_id = ? AND status = ?", productID, "approved").
		Scan(&result).Error
	return result.AvgRating, result.Count, err
}
