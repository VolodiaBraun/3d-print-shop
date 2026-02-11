package mock

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/brown/3d-print-shop/internal/delivery"
	"github.com/brown/3d-print-shop/internal/domain"
)

// MockProvider implements delivery.Provider using delivery zones from DB.
type MockProvider struct {
	zoneRepo domain.DeliveryZoneRepository
}

// New creates a new mock delivery provider.
func New(zoneRepo domain.DeliveryZoneRepository) *MockProvider {
	return &MockProvider{zoneRepo: zoneRepo}
}

func (p *MockProvider) Name() string { return "mock" }

func (p *MockProvider) CalculateDelivery(ctx context.Context, input delivery.CalculateInput) ([]delivery.DeliveryOption, error) {
	zone, err := p.zoneRepo.FindByCity(ctx, input.City)
	if err != nil {
		return nil, domain.ErrDeliveryNotAvailable
	}

	cost := zone.DeliveryCost
	isFree := false
	if zone.FreeDeliveryFrom != nil && input.OrderTotal >= *zone.FreeDeliveryFrom {
		cost = 0
		isFree = true
	}

	// Weight surcharge: +50 RUB per extra kg above 1kg
	if !isFree && input.TotalWeight > 1000 {
		extraKg := math.Ceil((input.TotalWeight - 1000) / 1000)
		cost += extraKg * 50
	}

	cost = math.Round(cost*100) / 100

	options := []delivery.DeliveryOption{
		{
			Type:             "courier",
			Name:             fmt.Sprintf("Курьерская доставка (%s)", zone.Name),
			Cost:             cost,
			OriginalCost:     zone.DeliveryCost,
			EstimatedDaysMin: zone.EstimatedDaysMin,
			EstimatedDaysMax: zone.EstimatedDaysMax,
			IsFreeDelivery:   isFree,
			ProviderName:     "mock",
		},
	}

	return options, nil
}

func (p *MockProvider) CreateShipment(_ context.Context, _ delivery.ShipmentInput) (*delivery.ShipmentResult, error) {
	tracking := fmt.Sprintf("MOCK-%d-%04d", time.Now().Unix(), rand.Intn(10000))
	return &delivery.ShipmentResult{
		TrackingNumber: tracking,
		ProviderName:   "mock",
	}, nil
}

func (p *MockProvider) TrackShipment(_ context.Context, _ string) ([]delivery.TrackingEvent, error) {
	now := time.Now()
	return []delivery.TrackingEvent{
		{
			Status:    "created",
			Location:  "Склад отправителя",
			Timestamp: now.Add(-48 * time.Hour).Format(time.RFC3339),
			Message:   "Отправление создано",
		},
		{
			Status:    "in_transit",
			Location:  "Сортировочный центр",
			Timestamp: now.Add(-24 * time.Hour).Format(time.RFC3339),
			Message:   "Передано в доставку",
		},
	}, nil
}
