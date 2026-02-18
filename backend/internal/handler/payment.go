package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// RegisterWebhookRoute registers the public payment gateway webhook.
// Called on the root router (outside /api/v1) to match Nginx proxy config.
func (h *PaymentHandler) RegisterWebhookRoute(router *gin.Engine) {
	router.POST("/webhook/payment", h.HandleWebhook)
}

// RegisterMockRoutes registers the mock payment confirmation endpoint.
// Only used with the mock provider; replace with a real payment page later.
func (h *PaymentHandler) RegisterMockRoutes(v1 *gin.RouterGroup) {
	v1.GET("/payment/mock/:orderNumber", h.MockConfirm)
}

// RegisterAdminRoutes registers admin-only payment actions.
func (h *PaymentHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("/orders/:id/regenerate-payment", h.RegeneratePayment)
}

// HandleWebhook processes an incoming payment status notification from the gateway.
// Always returns HTTP 200 to prevent repeated delivery from the gateway.
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "reason": "cannot read body"})
		return
	}

	headers := make(map[string]string)
	for k := range c.Request.Header {
		headers[k] = c.GetHeader(k)
	}

	if err := h.paymentService.HandleWebhook(c.Request.Context(), body, headers); err != nil {
		// Return 200 even on error to stop the gateway from retrying indefinitely.
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// MockConfirm simulates a successful payment for the mock provider.
// Marks the order as paid and redirects the customer to the order page.
// GET /api/v1/payment/mock/:orderNumber
func (h *PaymentHandler) MockConfirm(c *gin.Context) {
	orderNumber := c.Param("orderNumber")

	if err := h.paymentService.MarkPaidByOrderNumber(c.Request.Context(), orderNumber); err != nil {
		response.Error(c, http.StatusBadRequest, "PAYMENT_ERROR", err.Error())
		return
	}

	// Redirect to the order page on the frontend.
	c.Redirect(http.StatusFound, "/order/"+orderNumber)
}

// RegeneratePayment creates a new payment link for an order (e.g. after expiry).
// POST /admin/orders/:id/regenerate-payment
func (h *PaymentHandler) RegeneratePayment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	paymentURL, err := h.paymentService.RegeneratePaymentLink(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.NotFound(c, "Заказ не найден")
			return
		}
		response.Error(c, http.StatusBadRequest, "PAYMENT_ERROR", err.Error())
		return
	}

	response.OK(c, gin.H{"paymentLink": paymentURL})
}
