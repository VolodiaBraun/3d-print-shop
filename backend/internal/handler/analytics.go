package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

// RegisterAdminRoutes registers analytics admin routes.
func (h *AnalyticsHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/analytics/dashboard", h.GetDashboard)
	rg.GET("/analytics/chart", h.GetChart)
}

// GetDashboard handles GET /api/v1/admin/analytics/dashboard
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	metrics, err := h.analyticsService.GetDashboardMetrics(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, metrics)
}

// GetChart handles GET /api/v1/admin/analytics/chart?period=month
func (h *AnalyticsHandler) GetChart(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var from time.Time
	switch period {
	case "week":
		from = today.AddDate(0, 0, -6)
	case "month":
		from = today.AddDate(0, -1, 0)
	case "quarter":
		from = today.AddDate(0, -3, 0)
	case "year":
		from = today.AddDate(-1, 0, 0)
	default:
		response.Error(c, http.StatusBadRequest, "INVALID_PERIOD", "Допустимые периоды: week, month, quarter, year")
		return
	}

	data, err := h.analyticsService.GetChartData(c.Request.Context(), from, today)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, data)
}
