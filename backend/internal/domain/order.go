package domain

import (
	"context"
	"errors"
	"time"
)

type Order struct {
	ID              int         `gorm:"primaryKey" json:"id"`
	OrderNumber     string      `gorm:"uniqueIndex;not null" json:"orderNumber"`
	UserID          *int        `json:"userId,omitempty"`
	Status          string      `gorm:"not null;default:new" json:"status"`
	Subtotal        float64     `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	DiscountAmount  float64     `gorm:"type:decimal(10,2);default:0" json:"discountAmount"`
	DeliveryCost    float64     `gorm:"type:decimal(10,2);default:0" json:"deliveryCost"`
	TotalPrice      float64     `gorm:"type:decimal(10,2);not null" json:"totalPrice"`
	PromoCode       *string     `json:"promoCode,omitempty"`
	DeliveryMethod  string      `gorm:"not null;default:pickup" json:"deliveryMethod"`
	DeliveryAddress *string     `json:"deliveryAddress,omitempty"`
	PaymentMethod   string      `gorm:"not null;default:card" json:"paymentMethod"`
	IsPaid          bool        `gorm:"default:false" json:"isPaid"`
	CustomerName    string      `gorm:"not null" json:"customerName"`
	CustomerPhone   string      `gorm:"not null" json:"customerPhone"`
	CustomerEmail   *string     `json:"customerEmail,omitempty"`
	Notes           *string     `json:"notes,omitempty"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

type OrderItem struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	OrderID    int       `gorm:"not null" json:"orderId"`
	ProductID  int       `gorm:"not null" json:"productId"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	UnitPrice  float64   `gorm:"type:decimal(10,2);not null" json:"unitPrice"`
	TotalPrice float64   `gorm:"type:decimal(10,2);not null" json:"totalPrice"`
	Product    Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

var (
	ErrOrderNotFound = errors.New("order not found")
)

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id int) (*Order, error)
	FindByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
	ListByUserID(ctx context.Context, userID int) ([]Order, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	NextOrderNumber(ctx context.Context) (string, error)
}
