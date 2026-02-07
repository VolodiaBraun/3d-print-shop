package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

// ProductHandler handles product HTTP endpoints.
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler creates a new product handler.
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// RegisterPublicRoutes registers public product routes.
func (h *ProductHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/products", h.List)
	rg.GET("/products/:slug", h.GetBySlug)
	rg.GET("/search/suggestions", h.SearchSuggestions)
}

// RegisterAdminRoutes registers admin product routes.
func (h *ProductHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	products := rg.Group("/products")
	products.POST("", h.Create)
	products.PUT("/:id", h.Update)
	products.DELETE("/:id", h.Delete)
}

// List handles GET /api/v1/products?page=1&limit=20&category=slug&min_price=100&max_price=5000&material=PLA,PETG&sort=price_asc
func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := domain.ProductFilter{
		CategorySlug: c.Query("category"),
		Search:       c.Query("search"),
		Sort:         c.Query("sort"),
		Page:         page,
		Limit:        limit,
	}

	if v := c.Query("min_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinPrice = &p
		}
	}
	if v := c.Query("max_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MaxPrice = &p
		}
	}
	if v := c.Query("material"); v != "" {
		filter.Materials = strings.Split(v, ",")
	}

	result, err := h.productService.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Paginated(c, result.Products, response.PaginationMeta{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// Create handles POST /api/v1/admin/products
func (h *ProductHandler) Create(c *gin.Context) {
	var input service.CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Название и цена обязательны"},
		})
		return
	}

	product, err := h.productService.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrProductSlugExists) {
			response.Conflict(c, "Товар с таким slug уже существует")
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, product)
}

// GetBySlug handles GET /api/v1/products/:slug
func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.productService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			response.NotFound(c, "Товар не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, product)
}

// Update handles PUT /api/v1/admin/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	var input service.UpdateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Некорректные данные"},
		})
		return
	}

	product, err := h.productService.Update(c.Request.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			response.NotFound(c, "Товар не найден")
		case errors.Is(err, domain.ErrProductSlugExists):
			response.Conflict(c, "Товар с таким slug уже существует")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, product)
}

// Delete handles DELETE /api/v1/admin/products/:id (soft delete)
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	err = h.productService.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			response.NotFound(c, "Товар не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}

// SearchSuggestions handles GET /api/v1/search/suggestions?q=...
func (h *ProductHandler) SearchSuggestions(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.OK(c, []string{})
		return
	}

	suggestions, err := h.productService.SearchSuggestions(c.Request.Context(), q, 5)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, suggestions)
}
