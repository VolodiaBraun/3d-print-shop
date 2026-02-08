package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
)

type ValidatePromoInput struct {
	Code       string  `json:"code" binding:"required"`
	OrderTotal float64 `json:"orderTotal" binding:"required,gt=0"`
}

type CreatePromoInput struct {
	Code           string  `json:"code" binding:"required,min=3,max=50"`
	DiscountType   string  `json:"discountType" binding:"required,oneof=percent fixed"`
	DiscountValue  float64 `json:"discountValue" binding:"required,gt=0"`
	MinOrderAmount float64 `json:"minOrderAmount"`
	MaxUses        *int    `json:"maxUses"`
	IsActive       *bool   `json:"isActive"`
	StartsAt       *string `json:"startsAt"`
	ExpiresAt      *string `json:"expiresAt"`
	Description    *string `json:"description"`
}

type UpdatePromoInput struct {
	Code           *string  `json:"code"`
	DiscountType   *string  `json:"discountType" binding:"omitempty,oneof=percent fixed"`
	DiscountValue  *float64 `json:"discountValue" binding:"omitempty,gt=0"`
	MinOrderAmount *float64 `json:"minOrderAmount"`
	MaxUses        *int     `json:"maxUses"`
	IsActive       *bool    `json:"isActive"`
	StartsAt       *string  `json:"startsAt"`
	ExpiresAt      *string  `json:"expiresAt"`
	Description    *string  `json:"description"`
}

type PromoService struct {
	promoRepo domain.PromoRepository
	log       *zap.Logger
}

func NewPromoService(promoRepo domain.PromoRepository, log *zap.Logger) *PromoService {
	return &PromoService{
		promoRepo: promoRepo,
		log:       log,
	}
}

func (s *PromoService) Validate(ctx context.Context, input ValidatePromoInput) (*domain.PromoValidationResult, error) {
	promo, err := s.promoRepo.FindByCode(ctx, input.Code)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	if !promo.IsActive {
		return nil, domain.ErrPromoInactive
	}

	if promo.StartsAt != nil && now.Before(*promo.StartsAt) {
		return nil, domain.ErrPromoNotStarted
	}

	if promo.ExpiresAt != nil && now.After(*promo.ExpiresAt) {
		return nil, domain.ErrPromoExpired
	}

	if promo.MaxUses != nil && promo.UsedCount >= *promo.MaxUses {
		return nil, domain.ErrPromoUsedUp
	}

	if promo.MinOrderAmount > 0 && input.OrderTotal < promo.MinOrderAmount {
		return nil, domain.ErrPromoMinAmount
	}

	var discountAmount float64
	switch promo.DiscountType {
	case "percent":
		discountAmount = input.OrderTotal * promo.DiscountValue / 100
	case "fixed":
		discountAmount = promo.DiscountValue
	}
	// Don't exceed order total
	if discountAmount > input.OrderTotal {
		discountAmount = input.OrderTotal
	}
	discountAmount = math.Round(discountAmount*100) / 100

	return &domain.PromoValidationResult{
		Valid:          true,
		Code:           promo.Code,
		DiscountType:   promo.DiscountType,
		DiscountValue:  promo.DiscountValue,
		DiscountAmount: discountAmount,
	}, nil
}

func (s *PromoService) List(ctx context.Context) ([]domain.PromoCode, error) {
	return s.promoRepo.List(ctx)
}

func (s *PromoService) GetByID(ctx context.Context, id int) (*domain.PromoCode, error) {
	return s.promoRepo.FindByID(ctx, id)
}

func (s *PromoService) Create(ctx context.Context, input CreatePromoInput) (*domain.PromoCode, error) {
	promo := &domain.PromoCode{
		Code:           strings.ToUpper(strings.TrimSpace(input.Code)),
		DiscountType:   input.DiscountType,
		DiscountValue:  input.DiscountValue,
		MinOrderAmount: input.MinOrderAmount,
		MaxUses:        input.MaxUses,
		IsActive:       true,
		Description:    input.Description,
	}

	if input.IsActive != nil {
		promo.IsActive = *input.IsActive
	}

	if input.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *input.StartsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid startsAt format: %w", err)
		}
		promo.StartsAt = &t
	}

	if input.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *input.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expiresAt format: %w", err)
		}
		promo.ExpiresAt = &t
	}

	if err := s.promoRepo.Create(ctx, promo); err != nil {
		return nil, fmt.Errorf("create promo code: %w", err)
	}

	return promo, nil
}

func (s *PromoService) Update(ctx context.Context, id int, input UpdatePromoInput) (*domain.PromoCode, error) {
	promo, err := s.promoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Code != nil {
		promo.Code = *input.Code
	}
	if input.DiscountType != nil {
		promo.DiscountType = *input.DiscountType
	}
	if input.DiscountValue != nil {
		promo.DiscountValue = *input.DiscountValue
	}
	if input.MinOrderAmount != nil {
		promo.MinOrderAmount = *input.MinOrderAmount
	}
	if input.MaxUses != nil {
		promo.MaxUses = input.MaxUses
	}
	if input.IsActive != nil {
		promo.IsActive = *input.IsActive
	}
	if input.Description != nil {
		promo.Description = input.Description
	}

	if input.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *input.StartsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid startsAt format: %w", err)
		}
		promo.StartsAt = &t
	}
	if input.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *input.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expiresAt format: %w", err)
		}
		promo.ExpiresAt = &t
	}

	if err := s.promoRepo.Update(ctx, promo); err != nil {
		return nil, fmt.Errorf("update promo code: %w", err)
	}

	return promo, nil
}

func (s *PromoService) Delete(ctx context.Context, id int) error {
	return s.promoRepo.Delete(ctx, id)
}

func (s *PromoService) IncrementUsedCount(ctx context.Context, id int) error {
	return s.promoRepo.IncrementUsedCount(ctx, id)
}
