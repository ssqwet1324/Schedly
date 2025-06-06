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

// NewRedisRepository - подключаемся к редису
func NewRedisRepository(addr, password string, db int) *RedisRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisRepository{client: rdb}
}

// generateTaskKey - создание ключа для таски в Redis
func generateTaskKey(id string) string {
	return fmt.Sprintf("task:notification:%s", id)
}

// TaskScheduleKey - ключ для отсортированного множества задач
func TaskScheduleKey() string {
	return "task_schedule"
}

// SaveTask - сохраняем таску в Redis по ключу
func (r *RedisRepository) SaveTask(ctx context.Context, task *entity.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	key := generateTaskKey(task.ID)
	return r.client.Set(ctx, key, data, 0).Err()
}

// ScheduleTask - добавляет ID задачи в ZSet с временем отправки
func (r *RedisRepository) ScheduleTask(ctx context.Context, taskID string, eventTime time.Time) error {
	return r.client.ZAdd(ctx, TaskScheduleKey(), redis.Z{
		Score:  float64(eventTime.Unix()),
		Member: generateTaskKey(taskID),
	}).Err()
}

// ImmediateTaskForWorker - получает ближайшую задачу из ZSet и загружает её в структуру
func (r *RedisRepository) ImmediateTaskForWorker(ctx context.Context, task *entity.Task) error {
	result, err := r.client.ZRangeWithScores(ctx, TaskScheduleKey(), 0, 0).Result()
	if err != nil {
		return fmt.Errorf("ошибка при получении ближайшей задачи: %w", err)
	}
	if len(result) == 0 {
		return nil
	}

	key := result[0].Member.(string)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("не удалось получить данные задачи по ключу %s: %w", key, err)
	}

	if err := json.Unmarshal(data, task); err != nil {
		return fmt.Errorf("ошибка при разборе задачи: %w", err)
	}

	return nil
}

func (r *RedisRepository) RemoveTopZSetKey(ctx context.Context) error {
	results, err := r.client.ZRange(ctx, TaskScheduleKey(), 0, 0).Result()
	if err != nil || len(results) == 0 {
		return err
	}
	return r.client.ZRem(ctx, TaskScheduleKey(), results[0]).Err()
}

// GetTaskByKey - получить задачу по ключу
func (r *RedisRepository) GetTaskByKey(ctx context.Context, key string) (*entity.Task, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var task entity.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// FullDeleteTask - удалить задачу по id из redis и очереди
func (r *RedisRepository) FullDeleteTask(ctx context.Context, taskID string) error {
	key := generateTaskKey(taskID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete task from redis: %s: %w", key, err)
	}
	if err := r.client.ZRem(ctx, TaskScheduleKey(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete task from ScheduleTask: %s: %w", key, err)
	}
	return nil
}
