package postgres

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/brown/3d-print-shop/internal/config"
)

func NewDB(cfg config.DBConfig, log *zap.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	db, err := gorm.Open(postgres.Open(cfg.URL), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("database connected",
		zap.Int("maxOpenConns", cfg.MaxConnections),
		zap.Int("maxIdleConns", cfg.MaxIdleConns),
		zap.Duration("connMaxLifetime", cfg.ConnMaxLifetime),
	)

	return db, nil
}

func CloseDB(db *gorm.DB, log *zap.Logger) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("failed to get sql.DB for closing", zap.Error(err))
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Error("failed to close database", zap.Error(err))
		return
	}
	log.Info("database connection closed")
}

// NewDBWithTimeout is a helper for cases where you need to retry connection.
func NewDBWithTimeout(cfg config.DBConfig, log *zap.Logger, timeout time.Duration) (*gorm.DB, error) {
	deadline := time.Now().Add(timeout)

	for {
		db, err := NewDB(cfg, log)
		if err == nil {
			return db, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("database connection timeout after %s: %w", timeout, err)
		}

		log.Warn("database not ready, retrying...", zap.Error(err))
		time.Sleep(2 * time.Second)
	}
}
