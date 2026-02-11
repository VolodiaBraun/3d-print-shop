package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type ReviewService struct {
	reviewRepo  domain.ReviewRepository
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	db          *gorm.DB
	log         *zap.Logger
}

func NewReviewService(
	reviewRepo domain.ReviewRepository,
	orderRepo domain.OrderRepository,
	productRepo domain.ProductRepository,
	db *gorm.DB,
	log *zap.Logger,
) *ReviewService {
	return &ReviewService{
		reviewRepo:  reviewRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		db:          db,
		log:         log,
	}
}

type CreateReviewInput struct {
	ProductID int    `json:"productId" binding:"required"`
	OrderID   int    `json:"orderId" binding:"required"`
	Rating    int    `json:"rating" binding:"required,min=1,max=5"`
	Comment   string `json:"comment"`
}

func (s *ReviewService) CreateReview(ctx context.Context, userID int, input CreateReviewInput) (*domain.Review, error) {
	// 1. Check already reviewed
	_, err := s.reviewRepo.FindByUserAndProduct(ctx, userID, input.ProductID)
	if err == nil {
		return nil, domain.ErrAlreadyReviewed
	}
	if !errors.Is(err, domain.ErrReviewNotFound) {
		return nil, fmt.Errorf("check existing review: %w", err)
	}

	// 2. Check order exists and belongs to user
	order, err := s.orderRepo.FindByID(ctx, input.OrderID)
	if err != nil {
		return nil, fmt.Errorf("find order: %w", err)
	}
	if order.UserID == nil || *order.UserID != userID {
		return nil, domain.ErrOrderNotDelivered
	}

	// 3. Check order status is delivered
	if order.Status != "delivered" {
		return nil, domain.ErrOrderNotDelivered
	}

	// 4. Check product is in the order
	found := false
	for _, item := range order.Items {
		if item.ProductID == input.ProductID {
			found = true
			break
		}
	}
	if !found {
		return nil, domain.ErrOrderNotDelivered
	}

	// 5. Create review
	review := &domain.Review{
		UserID:    userID,
		ProductID: input.ProductID,
		OrderID:   input.OrderID,
		Rating:    input.Rating,
		Comment:   input.Comment,
		Status:    "pending",
	}
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	s.log.Info("review created",
		zap.Int("reviewID", review.ID),
		zap.Int("userID", userID),
		zap.Int("productID", input.ProductID),
		zap.Int("rating", input.Rating),
	)

	return review, nil
}

func (s *ReviewService) GetProductReviews(ctx context.Context, productID int) ([]domain.Review, error) {
	return s.reviewRepo.FindByProductID(ctx, productID)
}

func (s *ReviewService) GetMyReviews(ctx context.Context, userID int) ([]domain.Review, error) {
	return s.reviewRepo.ListByUserID(ctx, userID)
}

func (s *ReviewService) ApproveReview(ctx context.Context, id int) error {
	review, err := s.reviewRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.reviewRepo.UpdateStatus(ctx, id, "approved"); err != nil {
		return err
	}

	s.log.Info("review approved", zap.Int("reviewID", id))
	return s.recalculateProductRating(ctx, review.ProductID)
}

func (s *ReviewService) RejectReview(ctx context.Context, id int) error {
	review, err := s.reviewRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.reviewRepo.UpdateStatus(ctx, id, "rejected"); err != nil {
		return err
	}

	s.log.Info("review rejected", zap.Int("reviewID", id))

	// Recalculate only if it was previously approved
	if review.Status == "approved" {
		return s.recalculateProductRating(ctx, review.ProductID)
	}
	return nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, id int) error {
	review, err := s.reviewRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.reviewRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.log.Info("review deleted", zap.Int("reviewID", id))

	if review.Status == "approved" {
		return s.recalculateProductRating(ctx, review.ProductID)
	}
	return nil
}

func (s *ReviewService) ListReviews(ctx context.Context, filter domain.ReviewFilter) ([]domain.Review, int64, error) {
	return s.reviewRepo.List(ctx, filter)
}

func (s *ReviewService) recalculateProductRating(ctx context.Context, productID int) error {
	avg, count, err := s.reviewRepo.GetProductRatingStats(ctx, productID)
	if err != nil {
		s.log.Warn("failed to get rating stats", zap.Error(err))
		return nil
	}

	rounded := math.Round(avg*100) / 100

	err = s.db.WithContext(ctx).
		Model(&domain.Product{}).
		Where("id = ?", productID).
		Updates(map[string]interface{}{
			"rating":        rounded,
			"reviews_count": count,
		}).Error
	if err != nil {
		s.log.Warn("failed to update product rating", zap.Error(err))
	}
	return nil
}
