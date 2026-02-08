package domain

import (
	"context"
	"errors"
	"time"
)

type CartItem struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"not null" json:"-"`
	ProductID int       `gorm:"not null" json:"productId"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product"`
}

type Cart struct {
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"totalItems"`
	TotalPrice float64    `json:"totalPrice"`
}

var (
	ErrCartItemNotFound  = errors.New("cart item not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrProductInactive   = errors.New("product is not active")
)

type CartRepository interface {
	GetByUserID(ctx context.Context, userID int) ([]CartItem, error)
	FindItem(ctx context.Context, userID int, productID int) (*CartItem, error)
	FindItemByID(ctx context.Context, id int, userID int) (*CartItem, error)
	AddItem(ctx context.Context, item *CartItem) error
	UpdateQuantity(ctx context.Context, id int, userID int, quantity int) error
	RemoveItem(ctx context.Context, id int, userID int) error
	Clear(ctx context.Context, userID int) error
}
