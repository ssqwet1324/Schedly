package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"notif_service/internal/entity"
	"time"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(addr, password string, db int) *RedisRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return &RedisRepository{client: rdb}
}

// получить значение
func (r *RedisRepository) GetValue(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// изменить
func (r *RedisRepository) SetValue(ctx context.Context, key string, value string) error {
	return r.client.Set(ctx, key, value, 0).Err()
}
func (r *RedisRepository) SaveTask(ctx context.Context, task *entity.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	key := fmt.Sprintf("task:%s", task.ID)

	duration := time.Until(task.EventTime)
	if duration <= 0 {
		duration = time.Minute
	}
	err = r.client.Set(ctx, key, data, duration).Err()
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	return nil
}
