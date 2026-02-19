package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/cache"
	"github.com/brown/3d-print-shop/internal/domain"
)

const contentCacheTTL = 10 * time.Minute

type ContentService struct {
	repo  domain.ContentBlockRepository
	cache *cache.Store
	log   *zap.Logger
}

func NewContentService(repo domain.ContentBlockRepository, cache *cache.Store, log *zap.Logger) *ContentService {
	return &ContentService{repo: repo, cache: cache, log: log}
}

func contentCacheKey(slug string) string {
	return fmt.Sprintf("content:%s", slug)
}

// GetBlock returns content block data by slug. Uses Redis cache.
func (s *ContentService) GetBlock(ctx context.Context, slug string) (json.RawMessage, error) {
	// Try cache
	var cached json.RawMessage
	if found, err := s.cache.Get(ctx, contentCacheKey(slug), &cached); err == nil && found {
		return cached, nil
	}

	block, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("find content block %q: %w", slug, err)
	}

	_ = s.cache.Set(ctx, contentCacheKey(slug), block.Data, contentCacheTTL)
	return block.Data, nil
}

// UpdateBlock updates a content block's data and invalidates cache.
func (s *ContentService) UpdateBlock(ctx context.Context, slug string, data json.RawMessage) error {
	block := &domain.ContentBlock{
		Slug:      slug,
		Data:      data,
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Upsert(ctx, block); err != nil {
		return fmt.Errorf("upsert content block %q: %w", slug, err)
	}

	_ = s.cache.Delete(ctx, contentCacheKey(slug))

	s.log.Info("content block updated", zap.String("slug", slug))
	return nil
}
