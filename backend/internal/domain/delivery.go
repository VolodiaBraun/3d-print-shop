package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDeliveryZoneNotFound = errors.New("delivery zone not found")
	ErrPickupPointNotFound  = errors.New("pickup point not found")
	ErrDeliveryNotAvailable = errors.New("delivery not available for this city")
)

type DeliveryZone struct {
	ID               int       `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"not null" json:"name"`
	City             string    `gorm:"not null;index" json:"city"`
	Region           *string   `json:"region,omitempty"`
	DeliveryCost     float64   `gorm:"type:decimal(10,2);not null" json:"deliveryCost"`
	FreeDeliveryFrom *float64  `gorm:"type:decimal(10,2)" json:"freeDeliveryFrom,omitempty"`
	EstimatedDaysMin int       `gorm:"not null;default:1" json:"estimatedDaysMin"`
	EstimatedDaysMax int       `gorm:"not null;default:3" json:"estimatedDaysMax"`
	IsActive         bool      `gorm:"default:true" json:"isActive"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (DeliveryZone) TableName() string { return "delivery_zones" }

type PickupPoint struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Address      string    `gorm:"not null" json:"address"`
	City         string    `gorm:"not null;index" json:"city"`
	Latitude     *float64  `gorm:"type:decimal(10,7)" json:"latitude,omitempty"`
	Longitude    *float64  `gorm:"type:decimal(10,7)" json:"longitude,omitempty"`
	Phone        *string   `json:"phone,omitempty"`
	WorkingHours string    `gorm:"not null;default:'Пн-Пт 10:00-20:00'" json:"workingHours"`
	Provider     string    `gorm:"not null;default:'mock'" json:"provider"`
	ExternalID   *string   `json:"externalId,omitempty"`
	IsActive     bool      `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (PickupPoint) TableName() string { return "pickup_points" }

type DeliveryZoneRepository interface {
	FindByID(ctx context.Context, id int) (*DeliveryZone, error)
	FindByCity(ctx context.Context, city string) (*DeliveryZone, error)
	List(ctx context.Context) ([]DeliveryZone, error)
	Create(ctx context.Context, zone *DeliveryZone) error
	Update(ctx context.Context, zone *DeliveryZone) error
	Delete(ctx context.Context, id int) error
}

type PickupPointRepository interface {
	FindByID(ctx context.Context, id int) (*PickupPoint, error)
	FindByCity(ctx context.Context, city string) ([]PickupPoint, error)
	List(ctx context.Context) ([]PickupPoint, error)
	Create(ctx context.Context, point *PickupPoint) error
	Update(ctx context.Context, point *PickupPoint) error
	Delete(ctx context.Context, id int) error
}
