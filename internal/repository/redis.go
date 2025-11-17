package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis interface {
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type redisRepository struct {
	Rdb *redis.Client
}

func NewRedisRepository(rdb *redis.Client) Redis {
	return &redisRepository{
		rdb,
	}
}

func (r *redisRepository) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = r.Rdb.Set(ctx, key, jsonData, duration).Err()
	return err
}

func (r *redisRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *redisRepository) Del(ctx context.Context, key string) error {
	err := r.Rdb.Del(ctx, key)
	if err != nil {
		return errors.New("failed delete")
	}
	return nil
}
