package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCategoryHasChildren = errors.New("category has subcategories")
	ErrCategoryHasProducts = errors.New("category has products")
	ErrCategorySlugExists  = errors.New("category slug already exists")
)

// Category represents a product category.
type Category struct {
	ID           int         `gorm:"primaryKey" json:"id"`
	Name         string      `gorm:"not null" json:"name"`
	Slug         string      `gorm:"uniqueIndex;not null" json:"slug"`
	Description  *string     `json:"description,omitempty"`
	ParentID     *int        `json:"parentId,omitempty"`
	DisplayOrder int         `gorm:"default:0" json:"displayOrder"`
	ImageURL     *string     `json:"imageUrl,omitempty"`
	IsActive     bool        `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	Children     []Category  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// CategoryRepository defines the interface for category data access.
type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id int) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	FindAll(ctx context.Context) ([]Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id int) error
	HasChildren(ctx context.Context, id int) (bool, error)
	HasProducts(ctx context.Context, id int) (bool, error)
}
