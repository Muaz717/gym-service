package redis

import (
	"context"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/config"

	"github.com/redis/go-redis/v9"

	"net"
	"time"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(cfg config.Redis) (*Redis, error) {

	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DBRedis,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) Close() error {
	return r.client.Close()
}
