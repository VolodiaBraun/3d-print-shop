package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
)

// ProductService handles product business logic.
type ProductService struct {
	repo     domain.ProductRepository
	catRepo  domain.CategoryRepository
	log      *zap.Logger
}

// NewProductService creates a new product service.
func NewProductService(repo domain.ProductRepository, catRepo domain.CategoryRepository, log *zap.Logger) *ProductService {
	return &ProductService{repo: repo, catRepo: catRepo, log: log}
}

// CreateProductInput represents the input for creating a product.
type CreateProductInput struct {
	Name             string             `json:"name" binding:"required,min=1,max=255"`
	Slug             *string            `json:"slug"`
	Description      *string            `json:"description"`
	ShortDescription *string            `json:"shortDescription"`
	Price            float64            `json:"price" binding:"required,gt=0"`
	OldPrice         *float64           `json:"oldPrice"`
	StockQuantity    *int               `json:"stockQuantity"`
	SKU              *string            `json:"sku"`
	Weight           *float64           `json:"weight"`
	Dimensions       *domain.Dimensions `json:"dimensions"`
	Material         *string            `json:"material"`
	PrintTime        *int               `json:"printTime"`
	CategoryID       *int               `json:"categoryId"`
	IsFeatured       *bool              `json:"isFeatured"`
}

// UpdateProductInput represents the input for updating a product.
type UpdateProductInput struct {
	Name             *string            `json:"name" binding:"omitempty,min=1,max=255"`
	Slug             *string            `json:"slug"`
	Description      *string            `json:"description"`
	ShortDescription *string            `json:"shortDescription"`
	Price            *float64           `json:"price" binding:"omitempty,gt=0"`
	OldPrice         *float64           `json:"oldPrice"`
	StockQuantity    *int               `json:"stockQuantity"`
	SKU              *string            `json:"sku"`
	Weight           *float64           `json:"weight"`
	Dimensions       *domain.Dimensions `json:"dimensions"`
	Material         *string            `json:"material"`
	PrintTime        *int               `json:"printTime"`
	CategoryID       *int               `json:"categoryId"`
	IsActive         *bool              `json:"isActive"`
	IsFeatured       *bool              `json:"isFeatured"`
}

// Create creates a new product, auto-generating slug from name.
func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (*domain.Product, error) {
	slug := Slugify(input.Name)
	if input.Slug != nil && *input.Slug != "" {
		slug = NormalizeSlug(*input.Slug)
	}

	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	// Check slug uniqueness
	if _, err := s.repo.FindBySlug(ctx, slug); err == nil {
		return nil, domain.ErrProductSlugExists
	}

	// Validate category
	if input.CategoryID != nil {
		if _, err := s.catRepo.FindByID(ctx, *input.CategoryID); err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
				return nil, fmt.Errorf("category not found")
			}
			return nil, err
		}
	}

	stockQty := 0
	if input.StockQuantity != nil {
		stockQty = *input.StockQuantity
	}

	product := &domain.Product{
		Name:             input.Name,
		Slug:             slug,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		Price:            input.Price,
		OldPrice:         input.OldPrice,
		StockQuantity:    stockQty,
		SKU:              input.SKU,
		Weight:           input.Weight,
		Dimensions:       input.Dimensions,
		Material:         input.Material,
		PrintTime:        input.PrintTime,
		CategoryID:       input.CategoryID,
		IsActive:         true,
	}

	if input.IsFeatured != nil {
		product.IsFeatured = *input.IsFeatured
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	s.log.Info("product created", zap.Int("id", product.ID), zap.String("slug", product.Slug))
	return product, nil
}

// GetBySlug returns a product by slug with category info.
func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// GetByID returns a product by ID.
func (s *ProductService) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}

// Update updates an existing product.
func (s *ProductService) Update(ctx context.Context, id int, input UpdateProductInput) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.Slug != nil && *input.Slug != "" {
		newSlug := NormalizeSlug(*input.Slug)
		if newSlug != product.Slug {
			if existing, err := s.repo.FindBySlug(ctx, newSlug); err == nil && existing.ID != id {
				return nil, domain.ErrProductSlugExists
			}
			product.Slug = newSlug
		}
	}
	if input.Description != nil {
		product.Description = input.Description
	}
	if input.ShortDescription != nil {
		product.ShortDescription = input.ShortDescription
	}
	if input.Price != nil {
		product.Price = *input.Price
	}
	if input.OldPrice != nil {
		product.OldPrice = input.OldPrice
	}
	if input.StockQuantity != nil {
		product.StockQuantity = *input.StockQuantity
	}
	if input.SKU != nil {
		product.SKU = input.SKU
	}
	if input.Weight != nil {
		product.Weight = input.Weight
	}
	if input.Dimensions != nil {
		product.Dimensions = input.Dimensions
	}
	if input.Material != nil {
		product.Material = input.Material
	}
	if input.PrintTime != nil {
		product.PrintTime = input.PrintTime
	}
	if input.CategoryID != nil {
		if *input.CategoryID == 0 {
			product.CategoryID = nil
		} else {
			product.CategoryID = input.CategoryID
		}
	}
	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}
	if input.IsFeatured != nil {
		product.IsFeatured = *input.IsFeatured
	}

	product.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	s.log.Info("product updated", zap.Int("id", product.ID))
	return product, nil
}

// List returns a paginated, filtered list of products.
func (s *ProductService) List(ctx context.Context, filter domain.ProductFilter) (*domain.ProductListResult, error) {
	return s.repo.List(ctx, filter)
}

// Delete soft-deletes a product (sets is_active=false).
func (s *ProductService) Delete(ctx context.Context, id int) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("soft delete product: %w", err)
	}

	s.log.Info("product soft-deleted", zap.Int("id", id))
	return nil
}
