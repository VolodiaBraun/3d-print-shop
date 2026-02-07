package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
)

// CategoryService handles category business logic.
type CategoryService struct {
	repo domain.CategoryRepository
	log  *zap.Logger
}

// NewCategoryService creates a new category service.
func NewCategoryService(repo domain.CategoryRepository, log *zap.Logger) *CategoryService {
	return &CategoryService{repo: repo, log: log}
}

// CreateCategoryInput represents the input for creating a category.
type CreateCategoryInput struct {
	Name         string  `json:"name" binding:"required,min=1,max=255"`
	Slug         *string `json:"slug"`
	Description  *string `json:"description"`
	ParentID     *int    `json:"parentId"`
	DisplayOrder *int    `json:"displayOrder"`
	ImageURL     *string `json:"imageUrl"`
}

// UpdateCategoryInput represents the input for updating a category.
type UpdateCategoryInput struct {
	Name         *string `json:"name" binding:"omitempty,min=1,max=255"`
	Slug         *string `json:"slug"`
	Description  *string `json:"description"`
	ParentID     *int    `json:"parentId"`
	DisplayOrder *int    `json:"displayOrder"`
	ImageURL     *string `json:"imageUrl"`
	IsActive     *bool   `json:"isActive"`
}

// Create creates a new category, auto-generating slug from name if not provided.
func (s *CategoryService) Create(ctx context.Context, input CreateCategoryInput) (*domain.Category, error) {
	slug := Slugify(input.Name)
	if input.Slug != nil && *input.Slug != "" {
		slug = NormalizeSlug(*input.Slug)
	}

	if slug == "" {
		return nil, fmt.Errorf("slug is required (cannot generate from name)")
	}

	// Check slug uniqueness
	if _, err := s.repo.FindBySlug(ctx, slug); err == nil {
		return nil, domain.ErrCategorySlugExists
	}

	// Validate parent exists
	if input.ParentID != nil {
		if _, err := s.repo.FindByID(ctx, *input.ParentID); err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
				return nil, fmt.Errorf("parent category not found")
			}
			return nil, err
		}
	}

	displayOrder := 0
	if input.DisplayOrder != nil {
		displayOrder = *input.DisplayOrder
	}

	cat := &domain.Category{
		Name:         input.Name,
		Slug:         slug,
		Description:  input.Description,
		ParentID:     input.ParentID,
		DisplayOrder: displayOrder,
		ImageURL:     input.ImageURL,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	s.log.Info("category created", zap.Int("id", cat.ID), zap.String("slug", cat.Slug))
	return cat, nil
}

// GetTree returns the full category tree.
func (s *CategoryService) GetTree(ctx context.Context) ([]domain.Category, error) {
	return s.repo.FindAll(ctx)
}

// GetByID returns a category by ID.
func (s *CategoryService) GetByID(ctx context.Context, id int) (*domain.Category, error) {
	return s.repo.FindByID(ctx, id)
}

// Update updates an existing category.
func (s *CategoryService) Update(ctx context.Context, id int, input UpdateCategoryInput) (*domain.Category, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		cat.Name = *input.Name
	}
	if input.Slug != nil && *input.Slug != "" {
		newSlug := NormalizeSlug(*input.Slug)
		if newSlug != cat.Slug {
			if existing, err := s.repo.FindBySlug(ctx, newSlug); err == nil && existing.ID != id {
				return nil, domain.ErrCategorySlugExists
			}
			cat.Slug = newSlug
		}
	}
	if input.Description != nil {
		cat.Description = input.Description
	}
	if input.ParentID != nil {
		if *input.ParentID == 0 {
			cat.ParentID = nil
		} else {
			if *input.ParentID == id {
				return nil, fmt.Errorf("category cannot be its own parent")
			}
			cat.ParentID = input.ParentID
		}
	}
	if input.DisplayOrder != nil {
		cat.DisplayOrder = *input.DisplayOrder
	}
	if input.ImageURL != nil {
		cat.ImageURL = input.ImageURL
	}
	if input.IsActive != nil {
		cat.IsActive = *input.IsActive
	}

	cat.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}

	s.log.Info("category updated", zap.Int("id", cat.ID))
	return cat, nil
}

// Delete deletes a category if it has no children and no products.
func (s *CategoryService) Delete(ctx context.Context, id int) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("check children: %w", err)
	}
	if hasChildren {
		return domain.ErrCategoryHasChildren
	}

	hasProducts, err := s.repo.HasProducts(ctx, id)
	if err != nil {
		return fmt.Errorf("check products: %w", err)
	}
	if hasProducts {
		return domain.ErrCategoryHasProducts
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	s.log.Info("category deleted", zap.Int("id", id))
	return nil
}

// Slugify transliterates Cyrillic and normalizes to a URL-friendly slug.
// "Фигурки" → "figurki", "3D Модели" → "3d-modeli"
func Slugify(s string) string {
	s = transliterate(s)
	return NormalizeSlug(s)
}

// NormalizeSlug cleans a slug string: lowercase, only [a-z0-9-].
var nonSlugChar = regexp.MustCompile(`[^a-z0-9-]+`)
var multiDash = regexp.MustCompile(`-{2,}`)

func NormalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlugChar.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

var cyrMap = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo",
	'Ж': "Zh", 'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M",
	'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
	'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Shch",
	'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
}

func transliterate(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if mapped, ok := cyrMap[r]; ok {
			b.WriteString(mapped)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
