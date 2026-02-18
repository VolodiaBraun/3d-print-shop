package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// CustomOrderDetails stores extra data for orders of type "custom".
// One-to-one with orders (order_id is unique).
type CustomOrderDetails struct {
	ID                int             `gorm:"primaryKey" json:"id"`
	OrderID           int             `gorm:"uniqueIndex;not null" json:"orderId"`
	ClientDescription *string         `json:"clientDescription,omitempty"`
	AdminNotes        *string         `json:"adminNotes,omitempty"`
	// FileURLs: JSON array of S3 URLs for uploaded 3D model files (STL/OBJ/etc.)
	FileURLs      json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"fileUrls"`
	// PrintSettings: arbitrary JSON object with client preferences (material, color, layer height, etc.)
	PrintSettings json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"printSettings"`
	// Bitrix24 CRM integration fields (populated after bidirectional sync).
	BitrixDealID  *string `json:"bitrixDealId,omitempty"`
	BitrixStageID *string `json:"bitrixStageId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (CustomOrderDetails) TableName() string { return "custom_order_details" }

var (
	ErrCustomOrderNotFound        = errors.New("custom order details not found")
	ErrCustomOrderAlreadyConfirmed = errors.New("custom order is already confirmed")
	ErrOrderNotCustom             = errors.New("order is not a custom order")
)

type CustomOrderRepository interface {
	Create(ctx context.Context, details *CustomOrderDetails) error
	FindByOrderID(ctx context.Context, orderID int) (*CustomOrderDetails, error)
	Update(ctx context.Context, details *CustomOrderDetails) error
}
