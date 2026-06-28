package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{client: client}
}

func (r *RedisClient) SetJSON(ctx context.Context, key string, value any, ttlSeconds int) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	expiration := time.Duration(ttlSeconds) * time.Second
	return r.client.Set(ctx, key, payload, expiration).Err()
}

func (r *RedisClient) GetJSON(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if len(val) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}
