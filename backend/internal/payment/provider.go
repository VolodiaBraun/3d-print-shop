package payment

import (
	"context"
	"time"
)

// CreatePaymentInput contains data needed to create a payment.
type CreatePaymentInput struct {
	OrderID       int
	OrderNumber   string
	Amount        float64
	Description   string
	CustomerEmail *string
	ReturnURL     string // URL to redirect after payment
}

// PaymentResult is returned after successfully creating a payment.
type PaymentResult struct {
	ProviderPaymentID string    // ID assigned by the payment provider
	PaymentURL        string    // URL the customer visits to pay
	ExpiresAt         time.Time // When the payment link expires
}

// PaymentStatus represents the current state of a payment.
type PaymentStatus struct {
	ProviderPaymentID string
	Status            string     // pending | succeeded | cancelled | refunded
	PaidAmount        float64
	PaidAt            *time.Time
}

// WebhookEvent is the parsed result of an incoming payment webhook.
type WebhookEvent struct {
	ProviderPaymentID string
	OrderID           int
	Status            string  // succeeded | cancelled | refunded
	Amount            float64
}

// Provider is the abstract interface for payment providers.
// Implement this for YooKassa, Tinkoff, CloudPayments, or any other gateway.
type Provider interface {
	// Name returns the provider identifier used in logs and DB.
	Name() string

	// CreatePayment creates a new payment and returns the payment URL.
	CreatePayment(ctx context.Context, input CreatePaymentInput) (*PaymentResult, error)

	// GetPaymentStatus fetches the current status of a payment.
	GetPaymentStatus(ctx context.Context, providerPaymentID string) (*PaymentStatus, error)

	// CancelPayment cancels a pending payment.
	CancelPayment(ctx context.Context, providerPaymentID string) error

	// ValidateWebhook parses and validates an incoming webhook from the provider.
	// Returns a WebhookEvent on success, or an error if the payload is invalid.
	ValidateWebhook(body []byte, headers map[string]string) (*WebhookEvent, error)
}
