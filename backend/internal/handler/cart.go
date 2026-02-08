package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/middleware"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	cart := rg.Group("/cart")
	cart.Use(authMiddleware)
	cart.GET("", h.GetCart)
	cart.POST("/items", h.AddItem)
	cart.PUT("/items/:id", h.UpdateItem)
	cart.DELETE("/items/:id", h.RemoveItem)
	cart.DELETE("", h.Clear)
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Требуется авторизация")
		return
	}

	cart, err := h.cartService.GetCart(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, cart)
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Требуется авторизация")
		return
	}

	var input service.AddToCartInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	cart, err := h.cartService.AddItem(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, cart)
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Требуется авторизация")
		return
	}

	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	var input service.UpdateCartItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	cart, err := h.cartService.UpdateItem(c.Request.Context(), userID, itemID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, cart)
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Требуется авторизация")
		return
	}

	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	cart, err := h.cartService.RemoveItem(c.Request.Context(), userID, itemID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, cart)
}

func (h *CartHandler) Clear(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Требуется авторизация")
		return
	}

	if err := h.cartService.Clear(c.Request.Context(), userID); err != nil {
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}

func (h *CartHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrCartItemNotFound):
		response.NotFound(c, "Элемент корзины не найден")
	case errors.Is(err, domain.ErrProductNotFound):
		response.NotFound(c, "Товар не найден")
	case errors.Is(err, domain.ErrInsufficientStock):
		response.Error(c, http.StatusBadRequest, "INSUFFICIENT_STOCK", "Недостаточно товара на складе")
	case errors.Is(err, domain.ErrProductInactive):
		response.Error(c, http.StatusBadRequest, "PRODUCT_INACTIVE", "Товар недоступен")
	default:
		response.InternalError(c)
	}
}
