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
