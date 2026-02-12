package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

type AnalyticsService struct {
	repo domain.AnalyticsRepository
	db   *gorm.DB
	log  *zap.Logger
}

func NewAnalyticsService(repo domain.AnalyticsRepository, db *gorm.DB, log *zap.Logger) *AnalyticsService {
	return &AnalyticsService{repo: repo, db: db, log: log}
}

// GetDashboardMetrics returns all dashboard metrics.
func (s *AnalyticsService) GetDashboardMetrics(ctx context.Context) (*domain.DashboardMetrics, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Today - live data
	todayRevenue, todayOrders, _, err := s.repo.GetLiveRevenueForDate(ctx, today)
	if err != nil {
		s.log.Error("failed to get today revenue", zap.Error(err))
		return nil, err
	}
	todayCustomers, err := s.repo.GetNewCustomersForDate(ctx, today)
	if err != nil {
		s.log.Error("failed to get today customers", zap.Error(err))
		return nil, err
	}

	todayAvgCheck := 0.0
	if todayOrders > 0 {
		todayAvgCheck = todayRevenue / float64(todayOrders)
	}

	// This week (Mon–today)
	weekStart := today.AddDate(0, 0, -int(today.Weekday()-time.Monday+7)%7)
	weekRevenue, weekOrders, err := s.repo.GetRevenueByDateRange(ctx, weekStart, now)
	if err != nil {
		return nil, err
	}
	weekCustomers, err := s.repo.GetNewCustomersByDateRange(ctx, weekStart, now)
	if err != nil {
		return nil, err
	}
	weekAvg := 0.0
	if weekOrders > 0 {
		weekAvg = weekRevenue / float64(weekOrders)
	}

	// Previous week
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	prevWeekEnd := weekStart
	prevWeekRevenue, prevWeekOrders, err := s.repo.GetRevenueByDateRange(ctx, prevWeekStart, prevWeekEnd)
	if err != nil {
		return nil, err
	}
	prevWeekCustomers, err := s.repo.GetNewCustomersByDateRange(ctx, prevWeekStart, prevWeekEnd)
	if err != nil {
		return nil, err
	}
	prevWeekAvg := 0.0
	if prevWeekOrders > 0 {
		prevWeekAvg = prevWeekRevenue / float64(prevWeekOrders)
	}

	// This month
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthRevenue, monthOrders, err := s.repo.GetRevenueByDateRange(ctx, monthStart, now)
	if err != nil {
		return nil, err
	}
	monthCustomers, err := s.repo.GetNewCustomersByDateRange(ctx, monthStart, now)
	if err != nil {
		return nil, err
	}
	monthAvg := 0.0
	if monthOrders > 0 {
		monthAvg = monthRevenue / float64(monthOrders)
	}

	// Previous month
	prevMonthStart := monthStart.AddDate(0, -1, 0)
	prevMonthEnd := monthStart
	prevMonthRevenue, prevMonthOrders, err := s.repo.GetRevenueByDateRange(ctx, prevMonthStart, prevMonthEnd)
	if err != nil {
		return nil, err
	}
	prevMonthCustomers, err := s.repo.GetNewCustomersByDateRange(ctx, prevMonthStart, prevMonthEnd)
	if err != nil {
		return nil, err
	}
	prevMonthAvg := 0.0
	if prevMonthOrders > 0 {
		prevMonthAvg = prevMonthRevenue / float64(prevMonthOrders)
	}

	// Top products, low stock, pending orders
	topProducts, err := s.repo.GetTopProducts(ctx, 5)
	if err != nil {
		s.log.Error("failed to get top products", zap.Error(err))
	}
	if topProducts == nil {
		topProducts = []domain.TopProduct{}
	}
	lowStock, err := s.repo.GetLowStockProducts(ctx, 5)
	if err != nil {
		s.log.Error("failed to get low stock", zap.Error(err))
	}
	if lowStock == nil {
		lowStock = []domain.LowStockProduct{}
	}
	pendingOrders, err := s.repo.GetPendingOrders(ctx, 24)
	if err != nil {
		s.log.Error("failed to get pending orders", zap.Error(err))
	}
	if pendingOrders == nil {
		pendingOrders = []domain.PendingOrder{}
	}

	metrics := &domain.DashboardMetrics{
		Revenue: domain.PeriodMetric{
			Today:       todayRevenue,
			Week:        weekRevenue,
			Month:       monthRevenue,
			PrevWeek:    prevWeekRevenue,
			PrevMonth:   prevMonthRevenue,
			WeekChange:  percentChange(weekRevenue, prevWeekRevenue),
			MonthChange: percentChange(monthRevenue, prevMonthRevenue),
		},
		OrdersCount: domain.PeriodMetric{
			Today:       float64(todayOrders),
			Week:        float64(weekOrders),
			Month:       float64(monthOrders),
			PrevWeek:    float64(prevWeekOrders),
			PrevMonth:   float64(prevMonthOrders),
			WeekChange:  percentChange(float64(weekOrders), float64(prevWeekOrders)),
			MonthChange: percentChange(float64(monthOrders), float64(prevMonthOrders)),
		},
		AvgCheck: domain.PeriodMetric{
			Today:       todayAvgCheck,
			Week:        weekAvg,
			Month:       monthAvg,
			PrevWeek:    prevWeekAvg,
			PrevMonth:   prevMonthAvg,
			WeekChange:  percentChange(weekAvg, prevWeekAvg),
			MonthChange: percentChange(monthAvg, prevMonthAvg),
		},
		NewCustomers: domain.PeriodMetric{
			Today:       float64(todayCustomers),
			Week:        float64(weekCustomers),
			Month:       float64(monthCustomers),
			PrevWeek:    float64(prevWeekCustomers),
			PrevMonth:   float64(prevMonthCustomers),
			WeekChange:  percentChange(float64(weekCustomers), float64(prevWeekCustomers)),
			MonthChange: percentChange(float64(monthCustomers), float64(prevMonthCustomers)),
		},
		TopProducts:   topProducts,
		LowStock:      lowStock,
		PendingOrders: pendingOrders,
	}

	return metrics, nil
}

// GetChartData returns chart data points for the given date range.
func (s *AnalyticsService) GetChartData(ctx context.Context, from, to time.Time) ([]domain.ChartDataPoint, error) {
	// Get aggregated stats from the table
	stats, err := s.repo.GetStatsByDateRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Build a map for quick lookup
	statMap := make(map[string]domain.SalesDailyStat, len(stats))
	for _, s := range stats {
		statMap[s.Date.Format("2006-01-02")] = s
	}

	// Also get live data for today if it falls in range
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayKey := today.Format("2006-01-02")

	var todayRevenue float64
	var todayOrders int
	if !today.Before(from) && !today.After(to) {
		todayRevenue, todayOrders, _, err = s.repo.GetLiveRevenueForDate(ctx, today)
		if err != nil {
			s.log.Error("failed to get live data for chart", zap.Error(err))
		}
	}

	// Generate a data point for every day in range
	var points []domain.ChartDataPoint
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		point := domain.ChartDataPoint{Date: key}

		if key == todayKey {
			point.Revenue = todayRevenue
			point.OrdersCount = todayOrders
		} else if stat, ok := statMap[key]; ok {
			point.Revenue = stat.Revenue
			point.OrdersCount = stat.OrdersCount
		}

		points = append(points, point)
	}

	return points, nil
}

// AggregateMissing fills in missing daily stats from the last aggregated date until yesterday.
func (s *AnalyticsService) AggregateMissing(ctx context.Context) error {
	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())

	lastDate, err := s.repo.GetLastAggregatedDate(ctx)
	if err != nil {
		return err
	}

	var startDate time.Time
	if lastDate == nil {
		// Start from 30 days ago if no data
		startDate = yesterday.AddDate(0, 0, -30)
	} else {
		startDate = lastDate.AddDate(0, 0, 1)
	}

	if startDate.After(yesterday) {
		return nil // already up to date
	}

	s.log.Info("aggregating daily stats",
		zap.String("from", startDate.Format("2006-01-02")),
		zap.String("to", yesterday.Format("2006-01-02")),
	)

	for d := startDate; !d.After(yesterday); d = d.AddDate(0, 0, 1) {
		revenue, ordersCount, itemsSold, err := s.repo.GetLiveRevenueForDate(ctx, d)
		if err != nil {
			s.log.Error("aggregate: failed revenue", zap.Error(err), zap.String("date", d.Format("2006-01-02")))
			continue
		}
		newCustomers, err := s.repo.GetNewCustomersForDate(ctx, d)
		if err != nil {
			s.log.Error("aggregate: failed customers", zap.Error(err), zap.String("date", d.Format("2006-01-02")))
			continue
		}

		avgCheck := 0.0
		if ordersCount > 0 {
			avgCheck = revenue / float64(ordersCount)
		}

		stat := &domain.SalesDailyStat{
			Date:         d,
			Revenue:      revenue,
			OrdersCount:  ordersCount,
			AvgCheck:     avgCheck,
			NewCustomers: newCustomers,
			ItemsSold:    itemsSold,
		}

		if err := s.repo.UpsertDailyStat(ctx, stat); err != nil {
			s.log.Error("aggregate: failed upsert", zap.Error(err), zap.String("date", d.Format("2006-01-02")))
		}
	}

	return nil
}

// StartBackgroundAggregation runs aggregation immediately and then every hour.
func (s *AnalyticsService) StartBackgroundAggregation(ctx context.Context) {
	s.log.Info("starting background stats aggregation")

	if err := s.AggregateMissing(ctx); err != nil {
		s.log.Error("initial aggregation failed", zap.Error(err))
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("stopping background aggregation")
			return
		case <-ticker.C:
			if err := s.AggregateMissing(ctx); err != nil {
				s.log.Error("periodic aggregation failed", zap.Error(err))
			}
		}
	}
}

func percentChange(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return ((current - previous) / previous) * 100
}
