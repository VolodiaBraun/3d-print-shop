package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type DeliveryHandler struct {
	deliveryService *service.DeliveryService
}

func NewDeliveryHandler(deliveryService *service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{deliveryService: deliveryService}
}

func (h *DeliveryHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	d := rg.Group("/delivery")
	d.POST("/calculate", h.Calculate)
	d.GET("/pickup-points", h.GetPickupPoints)
}

func (h *DeliveryHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	zones := rg.Group("/delivery-zones")
	zones.GET("", h.ListZones)
	zones.GET("/:id", h.GetZone)
	zones.POST("", h.CreateZone)
	zones.PUT("/:id", h.UpdateZone)
	zones.DELETE("/:id", h.DeleteZone)

	points := rg.Group("/pickup-points")
	points.GET("", h.ListPickupPoints)
	points.GET("/:id", h.GetPickupPoint)
	points.POST("", h.CreatePickupPoint)
	points.PUT("/:id", h.UpdatePickupPoint)
	points.DELETE("/:id", h.DeletePickupPoint)
}

// --- Public ---

func (h *DeliveryHandler) Calculate(c *gin.Context) {
	var input service.CalculateDeliveryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	result, err := h.deliveryService.Calculate(c.Request.Context(), input)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, result)
}

func (h *DeliveryHandler) GetPickupPoints(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Параметр city обязателен")
		return
	}

	points, err := h.deliveryService.GetPickupPoints(c.Request.Context(), city)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, points)
}

// --- Admin: Delivery Zones ---

func (h *DeliveryHandler) ListZones(c *gin.Context) {
	zones, err := h.deliveryService.ListZones(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, zones)
}

func (h *DeliveryHandler) GetZone(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	zone, err := h.deliveryService.GetZone(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDeliveryZoneNotFound) {
			response.NotFound(c, "Зона доставки не найдена")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, zone)
}

func (h *DeliveryHandler) CreateZone(c *gin.Context) {
	var input service.CreateDeliveryZoneInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	zone, err := h.deliveryService.CreateZone(c.Request.Context(), input)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Created(c, zone)
}

func (h *DeliveryHandler) UpdateZone(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	var input service.UpdateDeliveryZoneInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	zone, err := h.deliveryService.UpdateZone(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrDeliveryZoneNotFound) {
			response.NotFound(c, "Зона доставки не найдена")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, zone)
}

func (h *DeliveryHandler) DeleteZone(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	if err := h.deliveryService.DeleteZone(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrDeliveryZoneNotFound) {
			response.NotFound(c, "Зона доставки не найдена")
			return
		}
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}

// --- Admin: Pickup Points ---

func (h *DeliveryHandler) ListPickupPoints(c *gin.Context) {
	points, err := h.deliveryService.ListPickupPoints(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, points)
}

func (h *DeliveryHandler) GetPickupPoint(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	point, err := h.deliveryService.GetPickupPoint(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPickupPointNotFound) {
			response.NotFound(c, "Пункт выдачи не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, point)
}

func (h *DeliveryHandler) CreatePickupPoint(c *gin.Context) {
	var input service.CreatePickupPointInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	point, err := h.deliveryService.CreatePickupPoint(c.Request.Context(), input)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Created(c, point)
}

func (h *DeliveryHandler) UpdatePickupPoint(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	var input service.UpdatePickupPointInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: err.Error()},
		})
		return
	}

	point, err := h.deliveryService.UpdatePickupPoint(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrPickupPointNotFound) {
			response.NotFound(c, "Пункт выдачи не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, point)
}

func (h *DeliveryHandler) DeletePickupPoint(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный ID")
		return
	}

	if err := h.deliveryService.DeletePickupPoint(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrPickupPointNotFound) {
			response.NotFound(c, "Пункт выдачи не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}
