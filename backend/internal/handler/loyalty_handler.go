package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/middleware"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type LoyaltyHandler struct {
	loyaltyService *service.LoyaltyService
}

func NewLoyaltyHandler(loyaltyService *service.LoyaltyService) *LoyaltyHandler {
	return &LoyaltyHandler{loyaltyService: loyaltyService}
}

// RegisterProtectedRoutes registers loyalty routes that require authentication.
func (h *LoyaltyHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.GET("/me/referral", h.GetReferralInfo)
	users.POST("/me/referral/apply", h.ApplyReferralCode)
	users.GET("/me/bonuses", h.GetBonusHistory)
}

// RegisterAdminRoutes registers admin loyalty routes.
func (h *LoyaltyHandler) RegisterAdminRoutes(admin *gin.RouterGroup) {
	loyalty := admin.Group("/loyalty")
	loyalty.GET("/settings", h.GetSettings)
	loyalty.PUT("/settings", h.UpdateSettings)
}

// GetReferralInfo handles GET /api/v1/users/me/referral
func (h *LoyaltyHandler) GetReferralInfo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	info, err := h.loyaltyService.GetReferralInfo(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, info)
}

type applyReferralInput struct {
	Code string `json:"code" binding:"required"`
}

// ApplyReferralCode handles POST /api/v1/users/me/referral/apply
func (h *LoyaltyHandler) ApplyReferralCode(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	var input applyReferralInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Введите реферальный код"},
		})
		return
	}

	err := h.loyaltyService.ApplyReferralCode(c.Request.Context(), userID, input.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoyaltyInactive):
			response.Error(c, http.StatusBadRequest, "LOYALTY_INACTIVE", "Реферальная программа неактивна")
		case errors.Is(err, service.ErrAlreadyReferred):
			response.Error(c, http.StatusConflict, "ALREADY_REFERRED", "Вы уже использовали реферальный код")
		case errors.Is(err, service.ErrReferralCodeNotFound):
			response.Error(c, http.StatusNotFound, "CODE_NOT_FOUND", "Реферальный код не найден")
		case errors.Is(err, service.ErrSelfReferral):
			response.Error(c, http.StatusBadRequest, "SELF_REFERRAL", "Нельзя использовать свой реферальный код")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, gin.H{"message": "Реферальный код успешно применён"})
}

// GetBonusHistory handles GET /api/v1/users/me/bonuses
func (h *LoyaltyHandler) GetBonusHistory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	history, err := h.loyaltyService.GetBonusHistory(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}

	if history == nil {
		history = []domain.BonusTransaction{}
	}

	response.OK(c, history)
}

// GetSettings handles GET /api/v1/admin/loyalty/settings
func (h *LoyaltyHandler) GetSettings(c *gin.Context) {
	settings, err := h.loyaltyService.GetSettings(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, settings)
}

// UpdateSettings handles PUT /api/v1/admin/loyalty/settings
func (h *LoyaltyHandler) UpdateSettings(c *gin.Context) {
	var input service.UpdateSettingsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Некорректные данные"},
		})
		return
	}

	settings, err := h.loyaltyService.UpdateSettings(c.Request.Context(), input)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, settings)
}
