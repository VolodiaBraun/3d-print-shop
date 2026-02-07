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

// CategoryHandler handles category HTTP endpoints.
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler creates a new category handler.
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// RegisterPublicRoutes registers public category routes.
func (h *CategoryHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/categories", h.GetTree)
}

// RegisterAdminRoutes registers admin category routes.
func (h *CategoryHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	cats := rg.Group("/categories")
	cats.POST("", h.Create)
	cats.PUT("/:id", h.Update)
	cats.DELETE("/:id", h.Delete)
}

// Create handles POST /api/v1/admin/categories
func (h *CategoryHandler) Create(c *gin.Context) {
	var input service.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Название категории обязательно"},
		})
		return
	}

	cat, err := h.categoryService.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrCategorySlugExists) {
			response.Conflict(c, "Категория с таким slug уже существует")
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, cat)
}

// GetTree handles GET /api/v1/categories
func (h *CategoryHandler) GetTree(c *gin.Context) {
	categories, err := h.categoryService.GetTree(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, categories)
}

// Update handles PUT /api/v1/admin/categories/:id
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	var input service.UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Некорректные данные"},
		})
		return
	}

	cat, err := h.categoryService.Update(c.Request.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCategoryNotFound):
			response.NotFound(c, "Категория не найдена")
		case errors.Is(err, domain.ErrCategorySlugExists):
			response.Conflict(c, "Категория с таким slug уже существует")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, cat)
}

// Delete handles DELETE /api/v1/admin/categories/:id
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	err = h.categoryService.Delete(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCategoryNotFound):
			response.NotFound(c, "Категория не найдена")
		case errors.Is(err, domain.ErrCategoryHasChildren):
			response.Error(c, http.StatusConflict, "HAS_CHILDREN", "Категория содержит подкатегории")
		case errors.Is(err, domain.ErrCategoryHasProducts):
			response.Error(c, http.StatusConflict, "HAS_PRODUCTS", "Категория содержит товары")
		default:
			response.InternalError(c)
		}
		return
	}

	response.NoContent(c)
}
