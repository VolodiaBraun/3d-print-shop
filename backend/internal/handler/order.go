package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	orders := rg.Group("/orders")
	orders.POST("", h.Create)
	orders.GET("/:orderNumber", h.GetByOrderNumber)
}

func (h *OrderHandler) Create(c *gin.Context) {
	var input service.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, order)
}

func (h *OrderHandler) GetByOrderNumber(c *gin.Context) {
	orderNumber := c.Param("orderNumber")

	order, err := h.orderService.GetByOrderNumber(c.Request.Context(), orderNumber)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.NotFound(c, "Заказ не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, order)
}

func (h *OrderHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		response.NotFound(c, "Заказ не найден")
	case errors.Is(err, domain.ErrProductNotFound):
		response.Error(c, http.StatusBadRequest, "PRODUCT_NOT_FOUND", "Товар не найден")
	case errors.Is(err, domain.ErrProductInactive):
		response.Error(c, http.StatusBadRequest, "PRODUCT_INACTIVE", "Товар недоступен")
	case errors.Is(err, domain.ErrInsufficientStock):
		response.Error(c, http.StatusBadRequest, "INSUFFICIENT_STOCK", "Недостаточно товара на складе")
	case errors.Is(err, domain.ErrPromoNotFound):
		response.Error(c, http.StatusBadRequest, "PROMO_NOT_FOUND", "Промокод не найден")
	case errors.Is(err, domain.ErrPromoExpired):
		response.Error(c, http.StatusBadRequest, "PROMO_EXPIRED", "Срок действия промокода истёк")
	case errors.Is(err, domain.ErrPromoInactive):
		response.Error(c, http.StatusBadRequest, "PROMO_INACTIVE", "Промокод неактивен")
	case errors.Is(err, domain.ErrPromoMinAmount):
		response.Error(c, http.StatusBadRequest, "PROMO_MIN_AMOUNT", "Минимальная сумма заказа для промокода не достигнута")
	default:
		response.InternalError(c)
	}
}
