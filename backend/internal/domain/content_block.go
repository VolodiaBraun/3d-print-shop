package domain

import (
	"context"
	"encoding/json"
	"time"
)

type ContentBlock struct {
	ID        int             `gorm:"primaryKey" json:"id"`
	Slug      string          `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Data      json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"data"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (ContentBlock) TableName() string {
	return "content_blocks"
}

type ContentBlockRepository interface {
	FindBySlug(ctx context.Context, slug string) (*ContentBlock, error)
	Upsert(ctx context.Context, block *ContentBlock) error
}
