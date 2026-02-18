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

type CustomOrderHandler struct {
	customOrderService *service.CustomOrderService
}

func NewCustomOrderHandler(customOrderService *service.CustomOrderService) *CustomOrderHandler {
	return &CustomOrderHandler{customOrderService: customOrderService}
}

// RegisterPublicRoutes — маршруты для клиентского фронтенда (без авторизации).
func (h *CustomOrderHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/custom-orders", h.SubmitRequest)
	rg.POST("/custom-orders/:id/files", h.UploadModelFile)
}

// RegisterAdminRoutes — маршруты для admin-панели.
func (h *CustomOrderHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("/custom-orders", h.CreateByAdmin)
	rg.GET("/custom-orders", h.ListCustomOrders)
	rg.GET("/custom-orders/:id", h.GetByID)
	rg.POST("/custom-orders/:id/confirm", h.ConfirmRequest)
	rg.POST("/custom-orders/:id/send-payment", h.SendPaymentLink)
	rg.POST("/custom-orders/:id/mark-paid", h.MarkPaidManually)
	rg.PUT("/custom-orders/:id", h.UpdateAdminDetails)
	rg.POST("/custom-orders/:id/files", h.UploadModelFile)
	rg.DELETE("/custom-orders/:id/files", h.DeleteModelFile)
}

// SubmitRequest — POST /api/v1/custom-orders (публичный)
// Клиент оставляет заявку. Цена неизвестна, статус = new.
func (h *CustomOrderHandler) SubmitRequest(c *gin.Context) {
	var input service.SubmitCustomOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	order, err := h.customOrderService.SubmitRequest(c.Request.Context(), input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "CREATE_ERROR", err.Error())
		return
	}

	response.Created(c, order)
}

// CreateByAdmin — POST /admin/custom-orders
// Администратор создаёт заказ с готовыми позициями и ценами.
func (h *CustomOrderHandler) CreateByAdmin(c *gin.Context) {
	var input service.CreateCustomOrderByAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	order, err := h.customOrderService.CreateByAdmin(c.Request.Context(), input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "CREATE_ERROR", err.Error())
		return
	}

	response.Created(c, order)
}

// ListCustomOrders — GET /admin/custom-orders?status=&page=&limit=
func (h *CustomOrderHandler) ListCustomOrders(c *gin.Context) {
	filter := domain.OrderFilter{
		Status: c.Query("status"),
	}
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		filter.Limit = limit
	}

	orders, total, err := h.customOrderService.ListCustomOrders(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_ERROR", err.Error())
		return
	}

	response.OK(c, gin.H{
		"items": orders,
		"total": total,
	})
}

// GetByID — GET /admin/custom-orders/:id
func (h *CustomOrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	order, err := h.customOrderService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.NotFound(c, "Заказ не найден")
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_ERROR", err.Error())
		return
	}

	response.OK(c, order)
}

// ConfirmRequest — POST /admin/custom-orders/:id/confirm
// Администратор подтверждает заявку и выставляет финальную цену.
func (h *CustomOrderHandler) ConfirmRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	var body struct {
		TotalPrice float64 `json:"totalPrice" binding:"required,gt=0"`
		AdminNotes *string `json:"adminNotes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	order, err := h.customOrderService.ConfirmRequest(c.Request.Context(), id, body.TotalPrice, body.AdminNotes)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			response.NotFound(c, "Заказ не найден")
		case errors.Is(err, domain.ErrOrderNotCustom):
			response.Error(c, http.StatusBadRequest, "NOT_CUSTOM_ORDER", "Заказ не является индивидуальным")
		case errors.Is(err, domain.ErrCustomOrderAlreadyConfirmed):
			response.Error(c, http.StatusConflict, "ALREADY_CONFIRMED", "Заказ уже подтверждён")
		default:
			response.Error(c, http.StatusInternalServerError, "CONFIRM_ERROR", err.Error())
		}
		return
	}

	response.OK(c, order)
}

// SendPaymentLink — POST /admin/custom-orders/:id/send-payment
// Создаёт / пересоздаёт ссылку на оплату.
func (h *CustomOrderHandler) SendPaymentLink(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	order, err := h.customOrderService.SendPaymentLink(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.NotFound(c, "Заказ не найден")
			return
		}
		response.Error(c, http.StatusBadRequest, "PAYMENT_ERROR", err.Error())
		return
	}

	response.OK(c, gin.H{"paymentLink": order.PaymentLink})
}

// MarkPaidManually — POST /admin/custom-orders/:id/mark-paid
// Вручную отмечает заказ оплаченным (наличные / перевод).
func (h *CustomOrderHandler) MarkPaidManually(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	if err := h.customOrderService.MarkPaidManually(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.NotFound(c, "Заказ не найден")
			return
		}
		response.Error(c, http.StatusBadRequest, "MARK_PAID_ERROR", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Заказ отмечен как оплаченный"})
}

// UpdateAdminDetails — PUT /admin/custom-orders/:id
// Обновляет admin_notes, print_settings, file_urls, bitrix-поля, total_price.
func (h *CustomOrderHandler) UpdateAdminDetails(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	var input service.UpdateCustomOrderAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	order, err := h.customOrderService.UpdateAdminDetails(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.NotFound(c, "Заказ не найден")
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_ERROR", err.Error())
		return
	}

	response.OK(c, order)
}

// UploadModelFile — POST /api/v1/custom-orders/:id/files  (также /admin/custom-orders/:id/files)
// Загружает 3D-файл (STL/OBJ/3MF/STEP/ZIP) к заказу. Multipart form, поле "file".
func (h *CustomOrderHandler) UploadModelFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "NO_FILE", "Файл не загружен (поле 'file')")
		return
	}
	defer file.Close()

	publicURL, err := h.customOrderService.UploadModelFile(
		c.Request.Context(),
		id,
		header.Filename,
		file,
		header.Size,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTooManyFiles):
			response.Error(c, http.StatusUnprocessableEntity, "TOO_MANY_FILES", err.Error())
		case errors.Is(err, service.ErrFileTooLarge):
			response.Error(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", err.Error())
		case errors.Is(err, service.ErrUnsupportedFormat):
			response.Error(c, http.StatusBadRequest, "UNSUPPORTED_FORMAT", err.Error())
		case errors.Is(err, domain.ErrCustomOrderNotFound):
			response.NotFound(c, "Заказ не найден")
		default:
			response.Error(c, http.StatusInternalServerError, "UPLOAD_ERROR", err.Error())
		}
		return
	}

	response.Created(c, gin.H{"url": publicURL, "fileName": header.Filename})
}

// DeleteModelFile — DELETE /admin/custom-orders/:id/files
// Удаляет файл по URL из заказа. Body: {"url": "https://..."}.
func (h *CustomOrderHandler) DeleteModelFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	var body struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.customOrderService.DeleteModelFile(c.Request.Context(), id, body.URL); err != nil {
		if errors.Is(err, domain.ErrCustomOrderNotFound) {
			response.NotFound(c, "Заказ не найден")
			return
		}
		response.Error(c, http.StatusBadRequest, "DELETE_ERROR", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Файл удалён"})
}
