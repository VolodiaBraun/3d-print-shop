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

type CreateOrderInput struct {
	Items           []OrderItemInput `json:"items" binding:"required,min=1,dive"`
	CustomerName    string           `json:"customerName" binding:"required"`
	CustomerPhone   string           `json:"customerPhone" binding:"required"`
	CustomerEmail   *string          `json:"customerEmail"`
	DeliveryMethod  string           `json:"deliveryMethod" binding:"required,oneof=pickup courier"`
	DeliveryAddress *string          `json:"deliveryAddress"`
	PaymentMethod   string           `json:"paymentMethod" binding:"required,oneof=card cash"`
	PromoCode       *string          `json:"promoCode"`
	Notes           *string          `json:"notes"`
	TelegramID      *int64           `json:"telegramId"`
}

type OrderItemInput struct {
	ProductID int `json:"productId" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,gt=0"`
}

type OrderService struct {
	orderRepo    domain.OrderRepository
	productRepo  domain.ProductRepository
	userRepo     domain.UserRepository
	promoService *PromoService
	db           *gorm.DB
	log          *zap.Logger
}

func NewOrderService(
	orderRepo domain.OrderRepository,
	productRepo domain.ProductRepository,
	userRepo domain.UserRepository,
	promoService *PromoService,
	db *gorm.DB,
	log *zap.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		productRepo:  productRepo,
		userRepo:     userRepo,
		promoService: promoService,
		db:           db,
		log:          log,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	// 1. Load and validate products
	productIDs := make([]int, len(input.Items))
	qtyMap := make(map[int]int)
	for i, item := range input.Items {
		productIDs[i] = item.ProductID
		qtyMap[item.ProductID] = item.Quantity
	}

	products, err := s.productRepo.FindByIDs(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("load products: %w", err)
	}

	productMap := make(map[int]*domain.Product)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	// Validate all products exist, are active, and have stock
	var subtotal float64
	orderItems := make([]domain.OrderItem, 0, len(input.Items))
	for _, item := range input.Items {
		p, ok := productMap[item.ProductID]
		if !ok {
			return nil, domain.ErrProductNotFound
		}
		if !p.IsActive {
			return nil, domain.ErrProductInactive
		}
		if p.StockQuantity < item.Quantity {
			return nil, domain.ErrInsufficientStock
		}

		itemTotal := math.Round(p.Price*float64(item.Quantity)*100) / 100
		subtotal += itemTotal

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  p.Price,
			TotalPrice: itemTotal,
		})
	}
	subtotal = math.Round(subtotal*100) / 100

	// 2. Validate promo code if provided
	var discountAmount float64
	var promoCode *string
	if input.PromoCode != nil && *input.PromoCode != "" {
		result, err := s.promoService.Validate(ctx, ValidatePromoInput{
			Code:       *input.PromoCode,
			OrderTotal: subtotal,
		})
		if err != nil {
			return nil, err
		}
		discountAmount = result.DiscountAmount
		promoCode = input.PromoCode
	}

	// 3. Calculate total
	totalPrice := math.Round((subtotal-discountAmount)*100) / 100
	if totalPrice < 0 {
		totalPrice = 0
	}

	// 4. Generate order number
	orderNumber, err := s.orderRepo.NextOrderNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate order number: %w", err)
	}

	// 5. Link Telegram user if telegramId is provided
	var userID *int
	if input.TelegramID != nil && *input.TelegramID != 0 {
		user, err := s.userRepo.FindByTelegramID(ctx, *input.TelegramID)
		if err != nil && errors.Is(err, domain.ErrUserNotFound) {
			// Create new user from Telegram data
			user = &domain.User{
				TelegramID: input.TelegramID,
				FirstName:  &input.CustomerName,
				Phone:      &input.CustomerPhone,
				Role:       "customer",
				IsActive:   true,
			}
			if createErr := s.userRepo.Create(ctx, user); createErr != nil {
				s.log.Warn("failed to create user from telegram", zap.Error(createErr))
			} else {
				userID = &user.ID
			}
		} else if err == nil {
			userID = &user.ID
			// Update phone if user didn't have one
			if user.Phone == nil || *user.Phone == "" {
				user.Phone = &input.CustomerPhone
				_ = s.userRepo.Update(ctx, user)
			}
		}
	}

	// 6. Create order in transaction
	order := &domain.Order{
		OrderNumber:     orderNumber,
		UserID:          userID,
		Status:          "new",
		Subtotal:        subtotal,
		DiscountAmount:  discountAmount,
		TotalPrice:      totalPrice,
		PromoCode:       promoCode,
		DeliveryMethod:  input.DeliveryMethod,
		DeliveryAddress: input.DeliveryAddress,
		PaymentMethod:   input.PaymentMethod,
		CustomerName:    input.CustomerName,
		CustomerPhone:   input.CustomerPhone,
		CustomerEmail:   input.CustomerEmail,
		Notes:           input.Notes,
		Items:           orderItems,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create order with items
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		// Decrease stock for each product
		for _, item := range input.Items {
			result := tx.Model(&domain.Product{}).
				Where("id = ? AND stock_quantity >= ?", item.ProductID, item.Quantity).
				UpdateColumn("stock_quantity", gorm.Expr("stock_quantity - ?", item.Quantity))
			if result.Error != nil {
				return fmt.Errorf("update stock: %w", result.Error)
			}
			if result.RowsAffected == 0 {
				return domain.ErrInsufficientStock
			}
		}

		// Increment promo code usage
		if promoCode != nil {
			if err := tx.Model(&domain.PromoCode{}).
				Where("UPPER(code) = UPPER(?)", *promoCode).
				UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
				s.log.Warn("failed to increment promo used_count", zap.Error(err))
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload order with preloaded relations
	created, err := s.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		return order, nil
	}

	s.log.Info("order created",
		zap.String("orderNumber", order.OrderNumber),
		zap.Float64("total", order.TotalPrice),
	)

	return created, nil
}

func (s *OrderService) GetByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	return s.orderRepo.FindByOrderNumber(ctx, orderNumber)
}

func (s *OrderService) GetByID(ctx context.Context, id int) (*domain.Order, error) {
	return s.orderRepo.FindByID(ctx, id)
}

func (s *OrderService) ListByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	return s.orderRepo.ListByUserID(ctx, userID)
}

var validTransitions = map[string][]string{
	"new":        {"confirmed", "cancelled"},
	"confirmed":  {"processing", "cancelled"},
	"processing": {"shipped"},
	"shipped":    {"delivered"},
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int, newStatus string) error {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	allowed := validTransitions[order.Status]
	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return domain.ErrOrderStatusInvalid
	}

	return s.orderRepo.UpdateStatus(ctx, id, newStatus)
}

func (s *OrderService) ListOrders(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int64, error) {
	return s.orderRepo.List(ctx, filter)
}

func (s *OrderService) UpdateTracking(ctx context.Context, id int, trackingNumber string) error {
	return s.orderRepo.UpdateTracking(ctx, id, trackingNumber)
}
