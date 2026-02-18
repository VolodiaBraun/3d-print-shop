// Package mock provides a mock payment provider for development and testing.
// It generates fake payment URLs pointing to a local confirmation endpoint.
// Replace with a real provider (YooKassa, Tinkoff, etc.) when going live.
package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/brown/3d-print-shop/internal/payment"
)

const defaultTTL = 24 * time.Hour

// Provider is the mock implementation of payment.Provider.
type Provider struct {
	appURL string // base URL of the application, e.g. https://avangard-print.ru
}

// New creates a mock payment provider.
// appURL is used to construct confirmation links, e.g. "https://avangard-print.ru".
func New(appURL string) *Provider {
	return &Provider{appURL: appURL}
}

func (p *Provider) Name() string { return "mock" }

func (p *Provider) CreatePayment(_ context.Context, input payment.CreatePaymentInput) (*payment.PaymentResult, error) {
	providerPaymentID := fmt.Sprintf("mock_%s_%d", input.OrderNumber, time.Now().UnixNano())
	// Points to a backend endpoint that simulates payment confirmation.
	paymentURL := fmt.Sprintf("%s/api/v1/payment/mock/%s", p.appURL, input.OrderNumber)

	return &payment.PaymentResult{
		ProviderPaymentID: providerPaymentID,
		PaymentURL:        paymentURL,
		ExpiresAt:         time.Now().Add(defaultTTL),
	}, nil
}

func (p *Provider) GetPaymentStatus(_ context.Context, providerPaymentID string) (*payment.PaymentStatus, error) {
	return &payment.PaymentStatus{
		ProviderPaymentID: providerPaymentID,
		Status:            "pending",
	}, nil
}

func (p *Provider) CancelPayment(_ context.Context, _ string) error {
	return nil
}

// ValidateWebhook is a no-op for the mock provider: real webhooks never arrive.
func (p *Provider) ValidateWebhook(_ []byte, _ map[string]string) (*payment.WebhookEvent, error) {
	return nil, fmt.Errorf("mock provider does not receive real payment webhooks")
}
