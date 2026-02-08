package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
)

type AddToCartInput struct {
	ProductID int `json:"productId" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemInput struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartService struct {
	cartRepo    domain.CartRepository
	productRepo domain.ProductRepository
	log         *zap.Logger
}

func NewCartService(cartRepo domain.CartRepository, productRepo domain.ProductRepository, log *zap.Logger) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		log:         log,
	}
}

func (s *CartService) GetCart(ctx context.Context, userID int) (*domain.Cart, error) {
	items, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	return s.buildCart(items), nil
}

func (s *CartService) AddItem(ctx context.Context, userID int, input AddToCartInput) (*domain.Cart, error) {
	product, err := s.productRepo.FindByID(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}

	if !product.IsActive {
		return nil, domain.ErrProductInactive
	}

	// Check if item already in cart
	existing, err := s.cartRepo.FindItem(ctx, userID, input.ProductID)
	if err != nil && !errors.Is(err, domain.ErrCartItemNotFound) {
		return nil, fmt.Errorf("find cart item: %w", err)
	}

	if existing != nil {
		// Update quantity
		newQty := existing.Quantity + input.Quantity
		if product.StockQuantity < newQty {
			return nil, domain.ErrInsufficientStock
		}
		if err := s.cartRepo.UpdateQuantity(ctx, existing.ID, userID, newQty); err != nil {
			return nil, fmt.Errorf("update cart item quantity: %w", err)
		}
	} else {
		// Add new item
		if product.StockQuantity < input.Quantity {
			return nil, domain.ErrInsufficientStock
		}
		item := &domain.CartItem{
			UserID:    userID,
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		}
		if err := s.cartRepo.AddItem(ctx, item); err != nil {
			return nil, fmt.Errorf("add cart item: %w", err)
		}
	}

	return s.GetCart(ctx, userID)
}

func (s *CartService) UpdateItem(ctx context.Context, userID int, itemID int, input UpdateCartItemInput) (*domain.Cart, error) {
	item, err := s.cartRepo.FindItemByID(ctx, itemID, userID)
	if err != nil {
		return nil, err
	}

	product, err := s.productRepo.FindByID(ctx, item.ProductID)
	if err != nil {
		return nil, err
	}

	if product.StockQuantity < input.Quantity {
		return nil, domain.ErrInsufficientStock
	}

	if err := s.cartRepo.UpdateQuantity(ctx, itemID, userID, input.Quantity); err != nil {
		return nil, fmt.Errorf("update cart item: %w", err)
	}

	return s.GetCart(ctx, userID)
}

func (s *CartService) RemoveItem(ctx context.Context, userID int, itemID int) (*domain.Cart, error) {
	if err := s.cartRepo.RemoveItem(ctx, itemID, userID); err != nil {
		return nil, err
	}
	return s.GetCart(ctx, userID)
}

func (s *CartService) Clear(ctx context.Context, userID int) error {
	if err := s.cartRepo.Clear(ctx, userID); err != nil {
		return fmt.Errorf("clear cart: %w", err)
	}
	return nil
}

func (s *CartService) buildCart(items []domain.CartItem) *domain.Cart {
	cart := &domain.Cart{
		Items: items,
	}
	for _, item := range items {
		cart.TotalItems += item.Quantity
		cart.TotalPrice += item.Product.Price * float64(item.Quantity)
	}
	// Round to 2 decimal places
	cart.TotalPrice = float64(int(cart.TotalPrice*100)) / 100
	return cart
}
