package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrProductSlugExists = errors.New("product slug already exists")
)

// Dimensions represents product dimensions in cm.
type Dimensions struct {
	Length float64 `json:"length,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

// Scan implements sql.Scanner for JSONB.
func (d *Dimensions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Dimensions: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, d)
}

// Value implements driver.Valuer for JSONB.
func (d Dimensions) Value() (driver.Value, error) {
	if d.Length == 0 && d.Width == 0 && d.Height == 0 {
		return nil, nil
	}
	return json.Marshal(d)
}

// Product represents a product in the catalog.
type Product struct {
	ID               int         `gorm:"primaryKey" json:"id"`
	Name             string      `gorm:"not null" json:"name"`
	Slug             string      `gorm:"uniqueIndex;not null" json:"slug"`
	Description      *string     `json:"description,omitempty"`
	ShortDescription *string     `json:"shortDescription,omitempty"`
	Price            float64     `gorm:"type:decimal(10,2);not null" json:"price"`
	OldPrice         *float64    `gorm:"type:decimal(10,2)" json:"oldPrice,omitempty"`
	StockQuantity    int         `gorm:"default:0" json:"stockQuantity"`
	SKU              *string     `gorm:"column:sku;uniqueIndex" json:"sku,omitempty"`
	Weight           *float64    `gorm:"type:decimal(10,2)" json:"weight,omitempty"`
	Dimensions       *Dimensions `gorm:"type:jsonb" json:"dimensions,omitempty"`
	Material         *string     `json:"material,omitempty"`
	PrintTime        *int        `json:"printTime,omitempty"`
	CategoryID       *int        `json:"categoryId,omitempty"`
	Category         *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images           []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
	IsActive         bool           `gorm:"default:true" json:"isActive"`
	IsFeatured       bool        `gorm:"default:false" json:"isFeatured"`
	ViewsCount       int         `gorm:"default:0" json:"viewsCount"`
	SalesCount       int         `gorm:"default:0" json:"salesCount"`
	Rating           float64     `gorm:"type:decimal(3,2);default:0" json:"rating"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

func (Product) TableName() string {
	return "products"
}

// ProductFilter defines filters for listing products.
type ProductFilter struct {
	CategorySlug    string   // filter by category slug (includes subcategories)
	MinPrice        *float64 // min price
	MaxPrice        *float64 // max price
	Materials       []string // filter by materials
	Search          string   // search in name/description
	Sort            string   // price_asc, price_desc, rating, newest, popular
	Page            int      // page number (1-based)
	Limit           int      // items per page
	IncludeInactive bool     // when true, includes inactive products (admin)
}

// ProductListResult contains paginated product list.
type ProductListResult struct {
	Products   []Product `json:"products"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"totalPages"`
}

// ProductRepository defines the interface for product data access.
type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id int) (*Product, error)
	FindBySlug(ctx context.Context, slug string) (*Product, error)
	List(ctx context.Context, filter ProductFilter) (*ProductListResult, error)
	Update(ctx context.Context, product *Product) error
	SoftDelete(ctx context.Context, id int) error
	FindByIDs(ctx context.Context, ids []int) ([]Product, error)
	SearchSuggestions(ctx context.Context, query string, limit int) ([]string, error)
}
