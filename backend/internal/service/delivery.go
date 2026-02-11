package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/delivery"
	"github.com/brown/3d-print-shop/internal/domain"
)

type CalculateDeliveryInput struct {
	City        string  `json:"city" binding:"required"`
	TotalWeight float64 `json:"totalWeight"`
	OrderTotal  float64 `json:"orderTotal" binding:"required,gt=0"`
}

type DeliveryCalculationResult struct {
	CourierOptions  []delivery.DeliveryOption `json:"courierOptions"`
	PickupPoints    []domain.PickupPoint      `json:"pickupPoints"`
	HasPickupPoints bool                      `json:"hasPickupPoints"`
}

type CreateDeliveryZoneInput struct {
	Name             string   `json:"name" binding:"required"`
	City             string   `json:"city" binding:"required"`
	Region           *string  `json:"region"`
	DeliveryCost     float64  `json:"deliveryCost" binding:"gte=0"`
	FreeDeliveryFrom *float64 `json:"freeDeliveryFrom"`
	EstimatedDaysMin int      `json:"estimatedDaysMin" binding:"required,gte=1"`
	EstimatedDaysMax int      `json:"estimatedDaysMax" binding:"required,gte=1"`
	IsActive         *bool    `json:"isActive"`
}

type UpdateDeliveryZoneInput struct {
	Name             *string  `json:"name"`
	City             *string  `json:"city"`
	Region           *string  `json:"region"`
	DeliveryCost     *float64 `json:"deliveryCost"`
	FreeDeliveryFrom *float64 `json:"freeDeliveryFrom"`
	EstimatedDaysMin *int     `json:"estimatedDaysMin"`
	EstimatedDaysMax *int     `json:"estimatedDaysMax"`
	IsActive         *bool    `json:"isActive"`
}

type CreatePickupPointInput struct {
	Name         string   `json:"name" binding:"required"`
	Address      string   `json:"address" binding:"required"`
	City         string   `json:"city" binding:"required"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Phone        *string  `json:"phone"`
	WorkingHours string   `json:"workingHours" binding:"required"`
	Provider     string   `json:"provider"`
	IsActive     *bool    `json:"isActive"`
}

type UpdatePickupPointInput struct {
	Name         *string  `json:"name"`
	Address      *string  `json:"address"`
	City         *string  `json:"city"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Phone        *string  `json:"phone"`
	WorkingHours *string  `json:"workingHours"`
	IsActive     *bool    `json:"isActive"`
}

type DeliveryService struct {
	provider        delivery.Provider
	zoneRepo        domain.DeliveryZoneRepository
	pickupPointRepo domain.PickupPointRepository
	log             *zap.Logger
}

func NewDeliveryService(
	provider delivery.Provider,
	zoneRepo domain.DeliveryZoneRepository,
	pickupPointRepo domain.PickupPointRepository,
	log *zap.Logger,
) *DeliveryService {
	return &DeliveryService{
		provider:        provider,
		zoneRepo:        zoneRepo,
		pickupPointRepo: pickupPointRepo,
		log:             log,
	}
}

// Calculate returns courier options and pickup points for a given city.
func (s *DeliveryService) Calculate(ctx context.Context, input CalculateDeliveryInput) (*DeliveryCalculationResult, error) {
	options, err := s.provider.CalculateDelivery(ctx, delivery.CalculateInput{
		City:        input.City,
		TotalWeight: input.TotalWeight,
		OrderTotal:  input.OrderTotal,
	})
	if err != nil {
		s.log.Debug("delivery calculation failed", zap.String("city", input.City), zap.Error(err))
		options = []delivery.DeliveryOption{}
	}

	points, err := s.pickupPointRepo.FindByCity(ctx, input.City)
	if err != nil {
		s.log.Debug("pickup points lookup failed", zap.String("city", input.City), zap.Error(err))
		points = []domain.PickupPoint{}
	}

	return &DeliveryCalculationResult{
		CourierOptions:  options,
		PickupPoints:    points,
		HasPickupPoints: len(points) > 0,
	}, nil
}

// GetPickupPoints returns active pickup points for a city.
func (s *DeliveryService) GetPickupPoints(ctx context.Context, city string) ([]domain.PickupPoint, error) {
	return s.pickupPointRepo.FindByCity(ctx, city)
}

// CalculateCourierCost returns the courier delivery cost for a city/total.
func (s *DeliveryService) CalculateCourierCost(ctx context.Context, city string, orderTotal float64, weight float64) (float64, string, error) {
	options, err := s.provider.CalculateDelivery(ctx, delivery.CalculateInput{
		City:        city,
		TotalWeight: weight,
		OrderTotal:  orderTotal,
	})
	if err != nil {
		return 0, "", err
	}
	if len(options) == 0 {
		return 0, "", domain.ErrDeliveryNotAvailable
	}
	opt := options[0]
	estimated := fmt.Sprintf("%d-%d дн.", opt.EstimatedDaysMin, opt.EstimatedDaysMax)
	return opt.Cost, estimated, nil
}

// --- Zone CRUD ---

func (s *DeliveryService) ListZones(ctx context.Context) ([]domain.DeliveryZone, error) {
	return s.zoneRepo.List(ctx)
}

func (s *DeliveryService) GetZone(ctx context.Context, id int) (*domain.DeliveryZone, error) {
	return s.zoneRepo.FindByID(ctx, id)
}

func (s *DeliveryService) CreateZone(ctx context.Context, input CreateDeliveryZoneInput) (*domain.DeliveryZone, error) {
	zone := &domain.DeliveryZone{
		Name:             input.Name,
		City:             input.City,
		Region:           input.Region,
		DeliveryCost:     input.DeliveryCost,
		FreeDeliveryFrom: input.FreeDeliveryFrom,
		EstimatedDaysMin: input.EstimatedDaysMin,
		EstimatedDaysMax: input.EstimatedDaysMax,
		IsActive:         true,
	}
	if input.IsActive != nil {
		zone.IsActive = *input.IsActive
	}
	if err := s.zoneRepo.Create(ctx, zone); err != nil {
		return nil, fmt.Errorf("create delivery zone: %w", err)
	}
	return zone, nil
}

func (s *DeliveryService) UpdateZone(ctx context.Context, id int, input UpdateDeliveryZoneInput) (*domain.DeliveryZone, error) {
	zone, err := s.zoneRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		zone.Name = *input.Name
	}
	if input.City != nil {
		zone.City = *input.City
	}
	if input.Region != nil {
		zone.Region = input.Region
	}
	if input.DeliveryCost != nil {
		zone.DeliveryCost = *input.DeliveryCost
	}
	if input.FreeDeliveryFrom != nil {
		zone.FreeDeliveryFrom = input.FreeDeliveryFrom
	}
	if input.EstimatedDaysMin != nil {
		zone.EstimatedDaysMin = *input.EstimatedDaysMin
	}
	if input.EstimatedDaysMax != nil {
		zone.EstimatedDaysMax = *input.EstimatedDaysMax
	}
	if input.IsActive != nil {
		zone.IsActive = *input.IsActive
	}
	if err := s.zoneRepo.Update(ctx, zone); err != nil {
		return nil, fmt.Errorf("update delivery zone: %w", err)
	}
	return zone, nil
}

func (s *DeliveryService) DeleteZone(ctx context.Context, id int) error {
	return s.zoneRepo.Delete(ctx, id)
}

// --- Pickup Point CRUD ---

func (s *DeliveryService) ListPickupPoints(ctx context.Context) ([]domain.PickupPoint, error) {
	return s.pickupPointRepo.List(ctx)
}

func (s *DeliveryService) GetPickupPoint(ctx context.Context, id int) (*domain.PickupPoint, error) {
	return s.pickupPointRepo.FindByID(ctx, id)
}

func (s *DeliveryService) CreatePickupPoint(ctx context.Context, input CreatePickupPointInput) (*domain.PickupPoint, error) {
	point := &domain.PickupPoint{
		Name:         input.Name,
		Address:      input.Address,
		City:         input.City,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		Phone:        input.Phone,
		WorkingHours: input.WorkingHours,
		Provider:     "mock",
		IsActive:     true,
	}
	if input.Provider != "" {
		point.Provider = input.Provider
	}
	if input.IsActive != nil {
		point.IsActive = *input.IsActive
	}
	if err := s.pickupPointRepo.Create(ctx, point); err != nil {
		return nil, fmt.Errorf("create pickup point: %w", err)
	}
	return point, nil
}

func (s *DeliveryService) UpdatePickupPoint(ctx context.Context, id int, input UpdatePickupPointInput) (*domain.PickupPoint, error) {
	point, err := s.pickupPointRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		point.Name = *input.Name
	}
	if input.Address != nil {
		point.Address = *input.Address
	}
	if input.City != nil {
		point.City = *input.City
	}
	if input.Latitude != nil {
		point.Latitude = input.Latitude
	}
	if input.Longitude != nil {
		point.Longitude = input.Longitude
	}
	if input.Phone != nil {
		point.Phone = input.Phone
	}
	if input.WorkingHours != nil {
		point.WorkingHours = *input.WorkingHours
	}
	if input.IsActive != nil {
		point.IsActive = *input.IsActive
	}
	if err := s.pickupPointRepo.Update(ctx, point); err != nil {
		return nil, fmt.Errorf("update pickup point: %w", err)
	}
	return point, nil
}

func (s *DeliveryService) DeletePickupPoint(ctx context.Context, id int) error {
	return s.pickupPointRepo.Delete(ctx, id)
}
