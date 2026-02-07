package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/config"
)

func NewRedis(cfg config.RedisConfig, log *zap.Logger) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	if cfg.Password != "" {
		opts.Password = cfg.Password
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Info("redis connected", zap.String("addr", opts.Addr))

	return client, nil
}
