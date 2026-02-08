package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

// ProductRepo implements domain.ProductRepository using GORM.
type ProductRepo struct {
	db *gorm.DB
}

// NewProductRepo creates a new product repository.
func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *ProductRepo) FindByID(ctx context.Context, id int) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_main DESC, display_order ASC")
		}).
		First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	return &product, err
}

func (r *ProductRepo) FindBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_main DESC, display_order ASC")
		}).
		Where("slug = ?", slug).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	return &product, err
}

func (r *ProductRepo) List(ctx context.Context, filter domain.ProductFilter) (*domain.ProductListResult, error) {
	query := r.db.WithContext(ctx).Model(&domain.Product{})
	if !filter.IncludeInactive {
		query = query.Where("is_active = true")
	}

	// Filter by category (including subcategories)
	if filter.CategorySlug != "" {
		var categoryIDs []int
		// Find category by slug, then collect all descendant IDs
		var cat domain.Category
		if err := r.db.Where("slug = ?", filter.CategorySlug).First(&cat).Error; err == nil {
			categoryIDs = append(categoryIDs, cat.ID)
			categoryIDs = append(categoryIDs, r.collectChildIDs(cat.ID)...)
			query = query.Where("category_id IN ?", categoryIDs)
		}
	}

	// Filter by price range
	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}

	// Filter by materials
	if len(filter.Materials) > 0 {
		query = query.Where("material IN ?", filter.Materials)
	}

	// Full-text search in name and description (Russian config)
	if filter.Search != "" {
		words := strings.Fields(filter.Search)
		tsQuery := strings.Join(words, " & ")
		query = query.Where("search_vector @@ to_tsquery('russian', ?)", tsQuery)
	}

	// Count total before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Sorting (relevance-first when searching without explicit sort)
	switch filter.Sort {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	case "rating":
		query = query.Order("rating DESC")
	case "newest":
		query = query.Order("created_at DESC")
	case "popular":
		query = query.Order("sales_count DESC")
	default:
		if filter.Search != "" {
			words := strings.Fields(filter.Search)
			tsQuery := strings.Join(words, " & ")
			query = query.Order(gorm.Expr("ts_rank(search_vector, to_tsquery('russian', ?)) DESC", tsQuery))
		} else {
			query = query.Order("created_at DESC")
		}
	}

	// Pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	var products []domain.Product
	err := query.
		Preload("Category").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC, id ASC")
		}).
		Offset(offset).
		Limit(filter.Limit).
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	return &domain.ProductListResult{
		Products:   products,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// collectChildIDs recursively collects all child category IDs.
func (r *ProductRepo) collectChildIDs(parentID int) []int {
	var children []domain.Category
	r.db.Where("parent_id = ?", parentID).Find(&children)

	var ids []int
	for _, child := range children {
		ids = append(ids, child.ID)
		ids = append(ids, r.collectChildIDs(child.ID)...)
	}
	return ids
}

func (r *ProductRepo) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *ProductRepo) SoftDelete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&domain.Product{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *ProductRepo) SearchSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	var names []string
	err := r.db.WithContext(ctx).
		Model(&domain.Product{}).
		Where("is_active = true AND name ILIKE ?", query+"%").
		Order("sales_count DESC, name ASC").
		Limit(limit).
		Pluck("name", &names).Error
	return names, err
}
