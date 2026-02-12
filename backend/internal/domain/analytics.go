package domain

import (
	"context"
	"time"
)

// SalesDailyStat represents aggregated daily sales statistics.
type SalesDailyStat struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Date         time.Time `gorm:"type:date;uniqueIndex;not null" json:"date"`
	Revenue      float64   `gorm:"type:decimal(12,2);not null;default:0" json:"revenue"`
	OrdersCount  int       `gorm:"not null;default:0" json:"ordersCount"`
	AvgCheck     float64   `gorm:"type:decimal(10,2);not null;default:0" json:"avgCheck"`
	NewCustomers int       `gorm:"not null;default:0" json:"newCustomers"`
	ItemsSold    int       `gorm:"not null;default:0" json:"itemsSold"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (SalesDailyStat) TableName() string {
	return "sales_daily_stats"
}

// PeriodMetric holds metrics for a single KPI across time periods.
type PeriodMetric struct {
	Today       float64 `json:"today"`
	Week        float64 `json:"week"`
	Month       float64 `json:"month"`
	PrevWeek    float64 `json:"prevWeek"`
	PrevMonth   float64 `json:"prevMonth"`
	WeekChange  float64 `json:"weekChange"`
	MonthChange float64 `json:"monthChange"`
}

// TopProduct is a top-selling product summary.
type TopProduct struct {
	ProductID int     `json:"productId"`
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	TotalSold int     `json:"totalSold"`
	Revenue   float64 `json:"revenue"`
	ImageURL  string  `json:"imageUrl,omitempty"`
}

// LowStockProduct is a product with low stock.
type LowStockProduct struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	StockQuantity int    `json:"stockQuantity"`
}

// PendingOrder is an order pending for too long.
type PendingOrder struct {
	ID           int       `json:"id"`
	OrderNumber  string    `json:"orderNumber"`
	TotalPrice   float64   `json:"totalPrice"`
	CreatedAt    time.Time `json:"createdAt"`
	HoursPending float64   `json:"hoursPending"`
}

// DashboardMetrics contains all dashboard data.
type DashboardMetrics struct {
	Revenue      PeriodMetric     `json:"revenue"`
	OrdersCount  PeriodMetric     `json:"ordersCount"`
	AvgCheck     PeriodMetric     `json:"avgCheck"`
	NewCustomers PeriodMetric     `json:"newCustomers"`
	TopProducts  []TopProduct     `json:"topProducts"`
	LowStock     []LowStockProduct `json:"lowStock"`
	PendingOrders []PendingOrder   `json:"pendingOrders"`
}

// ChartDataPoint is a single data point for the sales chart.
type ChartDataPoint struct {
	Date        string  `json:"date"`
	Revenue     float64 `json:"revenue"`
	OrdersCount int     `json:"ordersCount"`
}

// AnalyticsRepository defines data access for analytics.
type AnalyticsRepository interface {
	UpsertDailyStat(ctx context.Context, stat *SalesDailyStat) error
	GetLastAggregatedDate(ctx context.Context) (*time.Time, error)
	GetStatsByDateRange(ctx context.Context, from, to time.Time) ([]SalesDailyStat, error)

	GetLiveRevenueForDate(ctx context.Context, date time.Time) (revenue float64, ordersCount int, itemsSold int, err error)
	GetNewCustomersForDate(ctx context.Context, date time.Time) (int, error)

	GetRevenueByDateRange(ctx context.Context, from, to time.Time) (revenue float64, ordersCount int, err error)
	GetNewCustomersByDateRange(ctx context.Context, from, to time.Time) (int, error)

	GetTopProducts(ctx context.Context, limit int) ([]TopProduct, error)
	GetLowStockProducts(ctx context.Context, threshold int) ([]LowStockProduct, error)
	GetPendingOrders(ctx context.Context, olderThanHours int) ([]PendingOrder, error)
}
