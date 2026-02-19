package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type ContentHandler struct {
	contentService *service.ContentService
}

func NewContentHandler(contentService *service.ContentService) *ContentHandler {
	return &ContentHandler{contentService: contentService}
}

// RegisterPublicRoutes registers public content routes.
func (h *ContentHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/content/:slug", h.GetBlock)
}

// RegisterAdminRoutes registers admin content routes.
func (h *ContentHandler) RegisterAdminRoutes(admin *gin.RouterGroup) {
	admin.PUT("/content/:slug", h.UpdateBlock)
}

// GetBlock handles GET /api/v1/content/:slug
func (h *ContentHandler) GetBlock(c *gin.Context) {
	slug := c.Param("slug")

	data, err := h.contentService.GetBlock(c.Request.Context(), slug)
	if err != nil {
		response.NotFound(c, "Контентный блок не найден")
		return
	}

	c.JSON(http.StatusOK, json.RawMessage(data))
}

type updateBlockInput struct {
	Data json.RawMessage `json:"data" binding:"required"`
}

// UpdateBlock handles PUT /api/v1/admin/content/:slug
func (h *ContentHandler) UpdateBlock(c *gin.Context) {
	slug := c.Param("slug")

	var input updateBlockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Некорректные данные"},
		})
		return
	}

	if err := h.contentService.UpdateBlock(c.Request.Context(), slug, input.Data); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{"message": "Контент обновлён"})
}
