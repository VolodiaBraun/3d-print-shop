package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// Store provides typed caching operations over Redis.
type Store struct {
	client *redis.Client
}

// NewStore creates a new cache store.
func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

// Get retrieves a cached value by key. Returns false if not found.
func (s *Store) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, err
	}
	return true, nil
}

// Set stores a value with TTL.
func (s *Store) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key from cache.
func (s *Store) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// DeleteByPrefix removes all keys matching a prefix using SCAN.
func (s *Store) DeleteByPrefix(ctx context.Context, prefix string) error {
	iter := s.client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		s.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}
