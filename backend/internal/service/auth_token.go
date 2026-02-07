package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/pkg/jwt"
)

const refreshTokenPrefix = "refresh:"

// AuthTokenService manages JWT tokens with Redis-backed refresh token storage.
type AuthTokenService struct {
	jwt   *jwt.Manager
	redis *redis.Client
	log   *zap.Logger
}

// NewAuthTokenService creates a new auth token service.
func NewAuthTokenService(jwtManager *jwt.Manager, redisClient *redis.Client, log *zap.Logger) *AuthTokenService {
	return &AuthTokenService{
		jwt:   jwtManager,
		redis: redisClient,
		log:   log,
	}
}

// GenerateTokenPair creates a new access + refresh token pair and stores the refresh token in Redis.
func (s *AuthTokenService) GenerateTokenPair(ctx context.Context, userID int, role string) (*jwt.TokenPair, error) {
	pair, err := s.jwt.GenerateTokenPair(userID, role)
	if err != nil {
		return nil, fmt.Errorf("generate token pair: %w", err)
	}

	// Store refresh token in Redis: key = "refresh:<token>", value = "<userID>:<role>"
	key := refreshTokenPrefix + pair.RefreshToken
	val := fmt.Sprintf("%d:%s", userID, role)

	if err := s.redis.Set(ctx, key, val, s.jwt.RefreshExpiry()).Err(); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	s.log.Debug("token pair generated", zap.Int("userID", userID))
	return pair, nil
}

// RefreshTokenPair validates the old refresh token, invalidates it, and issues a new pair.
func (s *AuthTokenService) RefreshTokenPair(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	key := refreshTokenPrefix + refreshToken

	// Get and delete atomically — the token can only be used once.
	val, err := s.redis.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return nil, jwt.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	// Parse stored value "userID:role"
	var userID int
	var role string
	if _, err := fmt.Sscanf(val, "%d:%s", &userID, &role); err != nil {
		return nil, fmt.Errorf("parse refresh token data: %w", err)
	}

	// Generate new pair
	pair, err := s.GenerateTokenPair(ctx, userID, role)
	if err != nil {
		return nil, fmt.Errorf("generate new token pair: %w", err)
	}

	s.log.Debug("token pair refreshed", zap.Int("userID", userID))
	return pair, nil
}

// InvalidateRefreshToken removes a refresh token from Redis (e.g. on logout).
func (s *AuthTokenService) InvalidateRefreshToken(ctx context.Context, refreshToken string) error {
	key := refreshTokenPrefix + refreshToken
	deleted, err := s.redis.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	if deleted == 0 {
		return jwt.ErrInvalidToken
	}
	return nil
}

// InvalidateAllUserTokens removes all refresh tokens for a given user (e.g. password change).
func (s *AuthTokenService) InvalidateAllUserTokens(ctx context.Context, userID int) error {
	pattern := refreshTokenPrefix + "*"
	iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()

	prefix := fmt.Sprintf("%d:", userID)
	var keysToDelete []string

	for iter.Next(ctx) {
		key := iter.Val()
		val, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		if len(val) >= len(prefix) && val[:len(prefix)] == prefix {
			keysToDelete = append(keysToDelete, key)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan refresh tokens: %w", err)
	}

	if len(keysToDelete) > 0 {
		if err := s.redis.Del(ctx, keysToDelete...).Err(); err != nil {
			return fmt.Errorf("delete user refresh tokens: %w", err)
		}
		s.log.Info("invalidated user tokens", zap.Int("userID", userID), zap.Int("count", len(keysToDelete)))
	}

	return nil
}

// ValidateAccessToken delegates to the JWT manager.
func (s *AuthTokenService) ValidateAccessToken(tokenStr string) (*jwt.Claims, error) {
	return s.jwt.ValidateAccessToken(tokenStr)
}

// AccessExpiry returns the configured access token TTL.
func (s *AuthTokenService) AccessExpiry() time.Duration {
	return s.jwt.AccessExpiry()
}
