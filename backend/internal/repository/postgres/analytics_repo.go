package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/brown/3d-print-shop/internal/domain"
)

type AnalyticsRepo struct {
	db *gorm.DB
}

func NewAnalyticsRepo(db *gorm.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) UpsertDailyStat(ctx context.Context, stat *domain.SalesDailyStat) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"revenue", "orders_count", "avg_check", "new_customers", "items_sold",
			}),
		}).
		Create(stat).Error
}

func (r *AnalyticsRepo) GetLastAggregatedDate(ctx context.Context) (*time.Time, error) {
	var stat domain.SalesDailyStat
	err := r.db.WithContext(ctx).Order("date DESC").First(&stat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stat.Date, nil
}

func (r *AnalyticsRepo) GetStatsByDateRange(ctx context.Context, from, to time.Time) ([]domain.SalesDailyStat, error) {
	var stats []domain.SalesDailyStat
	err := r.db.WithContext(ctx).
		Where("date >= ? AND date <= ?", from, to).
		Order("date ASC").
		Find(&stats).Error
	return stats, err
}

func (r *AnalyticsRepo) GetLiveRevenueForDate(ctx context.Context, date time.Time) (float64, int, int, error) {
	var result struct {
		Revenue     float64
		OrdersCount int
		ItemsSold   int
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(o.total_price), 0) as revenue,
			COUNT(o.id) as orders_count,
			COALESCE(SUM(oi.quantity), 0) as items_sold
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.status != 'cancelled'
		  AND DATE(o.created_at) = ?
	`, date.Format("2006-01-02")).Scan(&result).Error
	return result.Revenue, result.OrdersCount, result.ItemsSold, err
}

func (r *AnalyticsRepo) GetNewCustomersForDate(ctx context.Context, date time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("role = 'customer' AND DATE(created_at) = ?", date.Format("2006-01-02")).
		Count(&count).Error
	return int(count), err
}

func (r *AnalyticsRepo) GetRevenueByDateRange(ctx context.Context, from, to time.Time) (float64, int, error) {
	var result struct {
		Revenue     float64
		OrdersCount int
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(total_price), 0) as revenue,
			COUNT(id) as orders_count
		FROM orders
		WHERE status != 'cancelled'
		  AND created_at >= ? AND created_at < ?
	`, from, to).Scan(&result).Error
	return result.Revenue, result.OrdersCount, err
}

func (r *AnalyticsRepo) GetNewCustomersByDateRange(ctx context.Context, from, to time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("role = 'customer' AND created_at >= ? AND created_at < ?", from, to).
		Count(&count).Error
	return int(count), err
}

func (r *AnalyticsRepo) GetTopProducts(ctx context.Context, limit int) ([]domain.TopProduct, error) {
	var products []domain.TopProduct
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			p.id as product_id,
			p.name,
			p.slug,
			COALESCE(SUM(oi.quantity), 0) as total_sold,
			COALESCE(SUM(oi.total_price), 0) as revenue,
			(SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.is_main = true LIMIT 1) as image_url
		FROM products p
		JOIN order_items oi ON oi.product_id = p.id
		JOIN orders o ON o.id = oi.order_id AND o.status != 'cancelled'
		GROUP BY p.id, p.name, p.slug
		ORDER BY total_sold DESC
		LIMIT ?
	`, limit).Scan(&products).Error
	return products, err
}

func (r *AnalyticsRepo) GetLowStockProducts(ctx context.Context, threshold int) ([]domain.LowStockProduct, error) {
	var products []domain.LowStockProduct
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, name, slug, stock_quantity
		FROM products
		WHERE stock_quantity <= ? AND is_active = true
		ORDER BY stock_quantity ASC
		LIMIT 10
	`, threshold).Scan(&products).Error
	return products, err
}

func (r *AnalyticsRepo) GetPendingOrders(ctx context.Context, olderThanHours int) ([]domain.PendingOrder, error) {
	var orders []domain.PendingOrder
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			id,
			order_number,
			total_price,
			created_at,
			EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600 as hours_pending
		FROM orders
		WHERE status = 'new'
		  AND created_at < NOW() - INTERVAL '1 hour' * ?
		ORDER BY created_at ASC
		LIMIT 10
	`, olderThanHours).Scan(&orders).Error
	return orders, err
}
