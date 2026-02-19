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

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterProtectedRoutes registers user routes that require authentication.
func (h *UserHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.GET("/me", h.GetProfile)
	users.PUT("/me", h.UpdateProfile)
	users.POST("/me/email/verify", h.SendVerificationCode)
	users.POST("/me/email/confirm", h.ConfirmVerificationCode)
}

// GetProfile handles GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.NotFound(c, "Пользователь не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, profile)
}

// UpdateProfile handles PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	var input service.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Некорректные данные"},
		})
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			response.Error(c, http.StatusConflict, "EMAIL_EXISTS", "Этот email уже используется")
		case errors.Is(err, domain.ErrUserNotFound):
			response.NotFound(c, "Пользователь не найден")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, profile)
}

// SendVerificationCode handles POST /api/v1/users/me/email/verify
func (h *UserHandler) SendVerificationCode(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	err := h.userService.SendVerificationCode(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoEmailToVerify):
			response.Error(c, http.StatusBadRequest, "NO_EMAIL", "Email не указан")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, gin.H{"message": "Код отправлен на вашу почту"})
}

type confirmCodeInput struct {
	Code string `json:"code" binding:"required"`
}

// ConfirmVerificationCode handles POST /api/v1/users/me/email/confirm
func (h *UserHandler) ConfirmVerificationCode(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	var input confirmCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Введите код подтверждения"},
		})
		return
	}

	err := h.userService.ConfirmVerificationCode(c.Request.Context(), userID, input.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidVerificationCode):
			response.Error(c, http.StatusBadRequest, "INVALID_CODE", "Неверный код")
		case errors.Is(err, service.ErrVerificationCodeExpired):
			response.Error(c, http.StatusBadRequest, "CODE_EXPIRED", "Код истёк, запросите новый")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, gin.H{"message": "Email подтверждён"})
}
