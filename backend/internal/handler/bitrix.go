package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/service"
)

// BitrixHandler processes incoming webhooks from Bitrix24.
type BitrixHandler struct {
	bitrixService *service.BitrixService
}

func NewBitrixHandler(bitrixService *service.BitrixService) *BitrixHandler {
	return &BitrixHandler{bitrixService: bitrixService}
}

// RegisterWebhookRoute registers the Bitrix24 incoming webhook endpoint.
// Called on the root router (outside /api/v1) to avoid JWT middleware.
func (h *BitrixHandler) RegisterWebhookRoute(router *gin.Engine) {
	router.POST("/webhook/bitrix", h.HandleEvent)
}

// HandleEvent processes an incoming Bitrix24 event notification.
// Bitrix sends a form-encoded POST with at minimum:
//
//	event=ONCRMDEALUPDATE
//	data[FIELDS][ID]=123
//
// We always respond 200 to prevent Bitrix from retrying.
// POST /webhook/bitrix
func (h *BitrixHandler) HandleEvent(c *gin.Context) {
	event := c.PostForm("event")
	dealID := c.PostForm("data[FIELDS][ID]")

	// Bitrix sometimes sends auth info, skip non-deal events silently
	if event != "ONCRMDEALUPDATE" && event != "ONCRMDEALADD" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "event": event})
		return
	}

	if err := h.bitrixService.HandleWebhook(c.Request.Context(), event, dealID); err != nil {
		// Always 200 to prevent Bitrix retries; log the error internally
		c.JSON(http.StatusOK, gin.H{"status": "error", "reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
