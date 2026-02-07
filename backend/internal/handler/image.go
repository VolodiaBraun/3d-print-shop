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

// ImageHandler handles product image HTTP endpoints.
type ImageHandler struct {
	imageService *service.ImageService
}

// NewImageHandler creates a new image handler.
func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

// RegisterAdminRoutes registers admin image routes.
func (h *ImageHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("/products/:id/images", h.Upload)
	rg.PUT("/products/:id/images/:imageId/main", h.SetMain)
	rg.DELETE("/products/images/:imageId", h.Delete)
}

// Upload handles POST /api/v1/admin/products/:id/images
func (h *ImageHandler) Upload(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID товара")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "NO_FILE", "Файл не загружен")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	img, err := h.imageService.Upload(c.Request.Context(), productID, file, contentType, header.Size)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			response.NotFound(c, "Товар не найден")
		default:
			response.Error(c, http.StatusBadRequest, "UPLOAD_ERROR", err.Error())
		}
		return
	}

	response.Created(c, img)
}

// SetMain handles PUT /api/v1/admin/products/:id/images/:imageId/main
func (h *ImageHandler) SetMain(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID товара")
		return
	}

	imageID, err := strconv.Atoi(c.Param("imageId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID изображения")
		return
	}

	err = h.imageService.SetMain(c.Request.Context(), productID, imageID)
	if err != nil {
		if errors.Is(err, domain.ErrImageNotFound) {
			response.NotFound(c, "Изображение не найдено")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{"message": "Главное изображение обновлено"})
}

// Delete handles DELETE /api/v1/admin/products/images/:imageId
func (h *ImageHandler) Delete(c *gin.Context) {
	imageID, err := strconv.Atoi(c.Param("imageId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID изображения")
		return
	}

	err = h.imageService.Delete(c.Request.Context(), imageID)
	if err != nil {
		if errors.Is(err, domain.ErrImageNotFound) {
			response.NotFound(c, "Изображение не найдено")
			return
		}
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}
