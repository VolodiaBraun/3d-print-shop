package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes registers auth routes on the given router group.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/telegram", h.TelegramLogin)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Email и пароль обязательны"},
		})
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			response.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный email или пароль")
		case errors.Is(err, domain.ErrAccountDisabled):
			response.Error(c, http.StatusForbidden, "ACCOUNT_DISABLED", "Аккаунт деактивирован")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, tokens)
}

// telegramLoginInput represents the Telegram login request body.
type telegramLoginInput struct {
	InitData string `json:"initData" binding:"required"`
}

// TelegramLogin handles POST /api/v1/auth/telegram
func (h *AuthHandler) TelegramLogin(c *gin.Context) {
	var input telegramLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "initData обязателен"},
		})
		return
	}

	resp, err := h.authService.LoginTelegram(c.Request.Context(), input.InitData)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInitData):
			response.Error(c, http.StatusUnauthorized, "INVALID_INIT_DATA", "Невалидные данные Telegram")
		case errors.Is(err, service.ErrInitDataExpired):
			response.Error(c, http.StatusUnauthorized, "INIT_DATA_EXPIRED", "Данные Telegram устарели")
		case errors.Is(err, domain.ErrAccountDisabled):
			response.Error(c, http.StatusForbidden, "ACCOUNT_DISABLED", "Аккаунт деактивирован")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, resp)
}

// refreshInput represents the refresh token request body.
type refreshInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var input refreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "refreshToken обязателен"},
		})
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Невалидный refresh token")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, tokens)
}
