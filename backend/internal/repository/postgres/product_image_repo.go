package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type ProductImageRepo struct {
	db *gorm.DB
}

func NewProductImageRepo(db *gorm.DB) *ProductImageRepo {
	return &ProductImageRepo{db: db}
}

func (r *ProductImageRepo) Create(ctx context.Context, image *domain.ProductImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *ProductImageRepo) FindByID(ctx context.Context, id int) (*domain.ProductImage, error) {
	var image domain.ProductImage
	err := r.db.WithContext(ctx).First(&image, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrImageNotFound
		}
		return nil, err
	}
	return &image, nil
}

func (r *ProductImageRepo) FindByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error) {
	var images []domain.ProductImage
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("display_order ASC, id ASC").
		Find(&images).Error
	return images, err
}

func (r *ProductImageRepo) SetMain(ctx context.Context, productID int, imageID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Reset all images for this product
		if err := tx.Model(&domain.ProductImage{}).
			Where("product_id = ?", productID).
			Update("is_main", false).Error; err != nil {
			return err
		}
		// Set the chosen image as main
		return tx.Model(&domain.ProductImage{}).
			Where("id = ? AND product_id = ?", imageID, productID).
			Update("is_main", true).Error
	})
}

func (r *ProductImageRepo) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&domain.ProductImage{}, id).Error
}

func (r *ProductImageRepo) CountByProductID(ctx context.Context, productID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ProductImage{}).
		Where("product_id = ?", productID).
		Count(&count).Error
	return count, err
}
