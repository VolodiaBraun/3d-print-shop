package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/payment"
)

// PaymentService handles payment link generation, webhook processing,
// and order payment status updates. It is provider-agnostic.
type PaymentService struct {
	provider  payment.Provider
	orderRepo domain.OrderRepository
	db        *gorm.DB
	log       *zap.Logger
	appURL    string // used to build return URLs and mock confirmation links
}

func NewPaymentService(
	provider payment.Provider,
	orderRepo domain.OrderRepository,
	db *gorm.DB,
	log *zap.Logger,
	appURL string,
) *PaymentService {
	return &PaymentService{
		provider:  provider,
		orderRepo: orderRepo,
		db:        db,
		log:       log,
		appURL:    appURL,
	}
}

// ProviderName returns the name of the active payment provider.
func (s *PaymentService) ProviderName() string {
	return s.provider.Name()
}

// InitiatePayment creates a payment for the given order, saves the payment link
// to the database, and returns the payment URL.
// Called automatically after order creation when payment_method == "card".
func (s *PaymentService) InitiatePayment(ctx context.Context, order *domain.Order) (string, error) {
	description := fmt.Sprintf("Заказ %s — АВАНГАРД 3D Print", order.OrderNumber)
	returnURL := fmt.Sprintf("%s/order/%s", s.appURL, order.OrderNumber)

	result, err := s.provider.CreatePayment(ctx, payment.CreatePaymentInput{
		OrderID:       order.ID,
		OrderNumber:   order.OrderNumber,
		Amount:        order.TotalPrice,
		Description:   description,
		CustomerEmail: order.CustomerEmail,
		ReturnURL:     returnURL,
	})
	if err != nil {
		return "", fmt.Errorf("create payment via %s: %w", s.provider.Name(), err)
	}

	providerName := s.provider.Name()
	if err := s.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", order.ID).
		Updates(map[string]interface{}{
			"payment_link":        result.PaymentURL,
			"payment_provider":    providerName,
			"payment_provider_id": result.ProviderPaymentID,
			"payment_expires_at":  result.ExpiresAt,
		}).Error; err != nil {
		return "", fmt.Errorf("save payment link to order %d: %w", order.ID, err)
	}

	s.log.Info("payment initiated",
		zap.String("orderNumber", order.OrderNumber),
		zap.String("provider", providerName),
		zap.String("providerPaymentID", result.ProviderPaymentID),
	)

	return result.PaymentURL, nil
}

// HandleWebhook parses and processes an incoming webhook from the payment provider.
// Marks the order as paid on success. Returns nil on ignored (non-success) events.
func (s *PaymentService) HandleWebhook(ctx context.Context, body []byte, headers map[string]string) error {
	event, err := s.provider.ValidateWebhook(body, headers)
	if err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}

	s.log.Info("payment webhook received",
		zap.String("providerPaymentID", event.ProviderPaymentID),
		zap.String("status", event.Status),
	)

	if event.Status != "succeeded" {
		// Log but don't fail — cancelled/refunded events are informational for now.
		return nil
	}

	var order domain.Order
	if err := s.db.WithContext(ctx).
		Where("payment_provider_id = ?", event.ProviderPaymentID).
		First(&order).Error; err != nil {
		return fmt.Errorf("order not found for provider payment %s: %w", event.ProviderPaymentID, err)
	}

	if order.IsPaid {
		s.log.Info("webhook: order already paid, skipping",
			zap.String("orderNumber", order.OrderNumber),
		)
		return nil
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", order.ID).
		Updates(map[string]interface{}{
			"is_paid":    true,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("mark order %d as paid: %w", order.ID, err)
	}

	s.log.Info("order marked as paid via webhook",
		zap.String("orderNumber", order.OrderNumber),
		zap.String("provider", s.provider.Name()),
	)

	return nil
}

// MarkPaidByOrderNumber marks an order as paid directly by its order number.
// Used by the mock confirmation endpoint and for manual cash payments.
func (s *PaymentService) MarkPaidByOrderNumber(ctx context.Context, orderNumber string) error {
	result := s.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("order_number = ? AND is_paid = false", orderNumber).
		Updates(map[string]interface{}{
			"is_paid":    true,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("mark order %s as paid: %w", orderNumber, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("order %s not found or already paid", orderNumber)
	}

	s.log.Info("order marked as paid manually",
		zap.String("orderNumber", orderNumber),
		zap.String("provider", s.provider.Name()),
	)
	return nil
}

// RegeneratePaymentLink cancels the old payment (if possible) and issues a new link.
// Useful when the previous link has expired.
func (s *PaymentService) RegeneratePaymentLink(ctx context.Context, orderID int) (string, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order.IsPaid {
		return "", fmt.Errorf("order %s is already paid", order.OrderNumber)
	}

	// Cancel existing payment at provider if one exists.
	if order.PaymentProviderID != nil && *order.PaymentProviderID != "" {
		if cancelErr := s.provider.CancelPayment(ctx, *order.PaymentProviderID); cancelErr != nil {
			s.log.Warn("failed to cancel old payment at provider",
				zap.String("orderNumber", order.OrderNumber),
				zap.Error(cancelErr),
			)
		}
	}

	return s.InitiatePayment(ctx, order)
}
