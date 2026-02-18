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
	// OrderType: "regular" (from catalog) | "custom" (individual 3D print order)
	OrderType       string      `gorm:"not null;default:regular" json:"orderType"`
	Status          string      `gorm:"not null;default:new" json:"status"`
	Subtotal        float64     `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	DiscountAmount  float64     `gorm:"type:decimal(10,2);default:0" json:"discountAmount"`
	DeliveryCost    float64     `gorm:"type:decimal(10,2);default:0" json:"deliveryCost"`
	BonusDiscount   float64     `gorm:"type:decimal(10,2);default:0" json:"bonusDiscount"`
	TotalPrice      float64     `gorm:"type:decimal(10,2);not null" json:"totalPrice"`
	PromoCode       *string     `json:"promoCode,omitempty"`
	DeliveryMethod  string      `gorm:"not null;default:pickup" json:"deliveryMethod"`
	DeliveryAddress *string     `json:"deliveryAddress,omitempty"`
	PaymentMethod   string      `gorm:"not null;default:card" json:"paymentMethod"`
	IsPaid          bool        `gorm:"default:false" json:"isPaid"`
	// Payment gateway fields (populated after InitiatePayment).
	PaymentLink       *string    `json:"paymentLink,omitempty"`
	PaymentProvider   *string    `json:"paymentProvider,omitempty"`
	PaymentProviderID *string    `gorm:"column:payment_provider_id" json:"-"` // internal, not exposed to clients
	PaymentExpiresAt  *time.Time `json:"paymentExpiresAt,omitempty"`
	CustomerName    string      `gorm:"not null" json:"customerName"`
	CustomerPhone   string      `gorm:"not null" json:"customerPhone"`
	CustomerEmail   *string     `json:"customerEmail,omitempty"`
	TrackingNumber    *string      `json:"trackingNumber,omitempty"`
	PickupPointID     *int         `json:"pickupPointId,omitempty"`
	PickupPoint       *PickupPoint `gorm:"foreignKey:PickupPointID" json:"pickupPoint,omitempty"`
	DeliveryProvider  *string      `json:"deliveryProvider,omitempty"`
	EstimatedDelivery *string      `json:"estimatedDelivery,omitempty"`
	Notes             *string      `json:"notes,omitempty"`
	Items           []OrderItem         `gorm:"foreignKey:OrderID" json:"items"`
	// CustomDetails is populated only for order_type == "custom".
	CustomDetails   *CustomOrderDetails `gorm:"foreignKey:OrderID" json:"customDetails,omitempty"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

type OrderItem struct {
	ID         int     `gorm:"primaryKey" json:"id"`
	OrderID    int     `gorm:"not null" json:"orderId"`
	// ProductID is nil for custom order line items (not from the catalog).
	ProductID  *int    `json:"productId,omitempty"`
	// Custom item fields — used when ProductID is nil.
	CustomItemName        *string `json:"customItemName,omitempty"`
	CustomItemDescription *string `json:"customItemDescription,omitempty"`
	Quantity   int      `gorm:"not null" json:"quantity"`
	UnitPrice  float64  `gorm:"type:decimal(10,2);not null" json:"unitPrice"`
	TotalPrice float64  `gorm:"type:decimal(10,2);not null" json:"totalPrice"`
	Product    *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type OrderFilter struct {
	Status    string
	OrderType string // "regular" | "custom" | "" (all)
	Page      int
	Limit     int
}

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderStatusInvalid  = errors.New("invalid status transition")
)

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id int) (*Order, error)
	FindByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
	List(ctx context.Context, filter OrderFilter) ([]Order, int64, error)
	ListByUserID(ctx context.Context, userID int) ([]Order, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateTracking(ctx context.Context, id int, trackingNumber string) error
	NextOrderNumber(ctx context.Context) (string, error)
}
