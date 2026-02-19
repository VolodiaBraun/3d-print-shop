package postgres

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/brown/3d-print-shop/internal/domain"
)

type ContentBlockRepo struct {
	db *gorm.DB
}

func NewContentBlockRepo(db *gorm.DB) *ContentBlockRepo {
	return &ContentBlockRepo{db: db}
}

func (r *ContentBlockRepo) FindBySlug(ctx context.Context, slug string) (*domain.ContentBlock, error) {
	var block domain.ContentBlock
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *ContentBlockRepo) Upsert(ctx context.Context, block *domain.ContentBlock) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{"data", "updated_at"}),
		}).
		Create(block).Error
}
