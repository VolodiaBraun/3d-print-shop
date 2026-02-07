package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrImageNotFound = errors.New("image not found")
)

// ProductImage represents an image associated with a product.
type ProductImage struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	ProductID    int       `gorm:"not null" json:"productId"`
	URL          string    `gorm:"not null" json:"url"`
	URLLarge     *string   `json:"urlLarge,omitempty"`
	URLMedium    *string   `json:"urlMedium,omitempty"`
	URLThumbnail *string   `json:"urlThumbnail,omitempty"`
	S3Key        string    `gorm:"not null" json:"s3Key"`
	IsMain       bool      `gorm:"default:false" json:"isMain"`
	DisplayOrder int       `gorm:"default:0" json:"displayOrder"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (ProductImage) TableName() string {
	return "product_images"
}

// ProductImageRepository defines the interface for product image data access.
type ProductImageRepository interface {
	Create(ctx context.Context, image *ProductImage) error
	FindByID(ctx context.Context, id int) (*ProductImage, error)
	FindByProductID(ctx context.Context, productID int) ([]ProductImage, error)
	SetMain(ctx context.Context, productID int, imageID int) error
	Delete(ctx context.Context, id int) error
	CountByProductID(ctx context.Context, productID int) (int64, error)
}
