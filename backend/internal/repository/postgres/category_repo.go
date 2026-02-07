package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

// CategoryRepo implements domain.CategoryRepository using GORM.
type CategoryRepo struct {
	db *gorm.DB
}

// NewCategoryRepo creates a new category repository.
func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *CategoryRepo) FindByID(ctx context.Context, id int) (*domain.Category, error) {
	var cat domain.Category
	err := r.db.WithContext(ctx).First(&cat, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	return &cat, err
}

func (r *CategoryRepo) FindBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var cat domain.Category
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&cat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	return &cat, err
}

// FindAll returns top-level categories with nested children (tree).
func (r *CategoryRepo) FindAll(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Order("display_order ASC, name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// Load children recursively
	for i := range categories {
		if err := r.loadChildren(ctx, &categories[i]); err != nil {
			return nil, err
		}
	}

	return categories, nil
}

func (r *CategoryRepo) loadChildren(ctx context.Context, parent *domain.Category) error {
	var children []domain.Category
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parent.ID).
		Order("display_order ASC, name ASC").
		Find(&children).Error
	if err != nil {
		return err
	}

	for i := range children {
		if err := r.loadChildren(ctx, &children[i]); err != nil {
			return err
		}
	}

	parent.Children = children
	return nil
}

func (r *CategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *CategoryRepo) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, id).Error
}

func (r *CategoryRepo) HasChildren(ctx context.Context, id int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Category{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *CategoryRepo) HasProducts(ctx context.Context, id int) (bool, error) {
	// Products table doesn't exist yet — will be updated in Story 2.3
	// For now, return false so categories can be deleted freely
	var count int64
	err := r.db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'products'").
		Scan(&count).Error
	if err != nil || count == 0 {
		return false, err
	}

	err = r.db.WithContext(ctx).
		Table("products").
		Where("category_id = ?", id).
		Count(&count).Error
	return count > 0, err
}
