package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrReviewNotFound   = errors.New("review not found")
	ErrAlreadyReviewed  = errors.New("already reviewed this product")
	ErrOrderNotDelivered = errors.New("order not delivered")
)

type Review struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"not null" json:"userId"`
	ProductID int       `gorm:"not null" json:"productId"`
	OrderID   int       `gorm:"not null" json:"orderId"`
	Rating    int       `gorm:"not null" json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	Status    string    `gorm:"default:pending" json:"status"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Review) TableName() string {
	return "reviews"
}

type ReviewFilter struct {
	Status    string
	ProductID int
	Page      int
	Limit     int
}

type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	FindByID(ctx context.Context, id int) (*Review, error)
	FindByProductID(ctx context.Context, productID int) ([]Review, error)
	FindByUserAndProduct(ctx context.Context, userID, productID int) (*Review, error)
	List(ctx context.Context, filter ReviewFilter) ([]Review, int64, error)
	ListByUserID(ctx context.Context, userID int) ([]Review, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
	GetProductRatingStats(ctx context.Context, productID int) (avgRating float64, count int, err error)
}
