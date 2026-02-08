package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type PromoHandler struct {
	promoService *service.PromoService
}

func NewPromoHandler(promoService *service.PromoService) *PromoHandler {
	return &PromoHandler{promoService: promoService}
}

func (h *PromoHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/promo/validate", h.Validate)
}

func (h *PromoHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	promos := rg.Group("/promo-codes")
	promos.GET("", h.List)
	promos.GET("/:id", h.GetByID)
	promos.POST("", h.Create)
	promos.PUT("/:id", h.Update)
	promos.DELETE("/:id", h.Delete)
}

func (h *PromoHandler) Validate(c *gin.Context) {
	var input service.ValidatePromoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	result, err := h.promoService.Validate(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *PromoHandler) List(c *gin.Context) {
	promos, err := h.promoService.List(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, promos)
}

func (h *PromoHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	promo, err := h.promoService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPromoNotFound) {
			response.NotFound(c, "Промокод не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, promo)
}

func (h *PromoHandler) Create(c *gin.Context) {
	var input service.CreatePromoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	promo, err := h.promoService.Create(c.Request.Context(), input)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Created(c, promo)
}

func (h *PromoHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	var input service.UpdatePromoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	promo, err := h.promoService.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrPromoNotFound) {
			response.NotFound(c, "Промокод не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, promo)
}

func (h *PromoHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	if err := h.promoService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrPromoNotFound) {
			response.NotFound(c, "Промокод не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}

func (h *PromoHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrPromoNotFound):
		response.NotFound(c, "Промокод не найден")
	case errors.Is(err, domain.ErrPromoInactive):
		response.Error(c, http.StatusBadRequest, "PROMO_INACTIVE", "Промокод неактивен")
	case errors.Is(err, domain.ErrPromoExpired):
		response.Error(c, http.StatusBadRequest, "PROMO_EXPIRED", "Срок действия промокода истёк")
	case errors.Is(err, domain.ErrPromoNotStarted):
		response.Error(c, http.StatusBadRequest, "PROMO_NOT_STARTED", "Промокод ещё не активен")
	case errors.Is(err, domain.ErrPromoUsedUp):
		response.Error(c, http.StatusBadRequest, "PROMO_USED_UP", "Лимит использований промокода исчерпан")
	case errors.Is(err, domain.ErrPromoMinAmount):
		response.Error(c, http.StatusBadRequest, "PROMO_MIN_AMOUNT", fmt.Sprintf("Минимальная сумма заказа для промокода не достигнута"))
	default:
		response.InternalError(c)
	}
}
