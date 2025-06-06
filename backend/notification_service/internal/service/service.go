package service

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"notif_service/internal/entity"
	"notif_service/internal/kafka"
	"notif_service/internal/repository"
)

type Service struct {
	Repo     repository.RedisRepository
	Producer *kafka.Producer
}

func NewService(repo repository.RedisRepository, producer *kafka.Producer) *Service {
	return &Service{Repo: repo, Producer: producer}
}

// CreateTask - создать задачу и добавить ее в планировщик
func (s *Service) CreateTask(ctx context.Context, e entity.Task) (entity.Task, error) {
	id := uuid.New().String()
	task := &entity.Task{
		ID:        id,
		Title:     e.Title,
		EventTime: e.EventTime,
		Body:      e.Body,
		Email:     e.Email,
	}

	// тут сохраняем задачу
	if err := s.Repo.SaveTask(ctx, task); err != nil {
		log.Printf("Failed to save task: %v", err)
		return entity.Task{}, err
	}

	// тут добавляем ее в планировщике по времени
	if err := s.Repo.ScheduleTask(ctx, task.ID, task.EventTime); err != nil {
		log.Printf("Failed to schedule task: %v", err)
		return entity.Task{}, err
	}

	data, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		return entity.Task{}, err
	}

	err = s.Producer.WriteMessages(ctx, task.ID, data)
	if err != nil {
		log.Printf("Failed to produce message: %v", err)
	}
	return *task, nil
}

// GetNearTask - получить ближайшую задачу
func (s *Service) GetNearTask(ctx context.Context) (entity.Task, error) {
	var task entity.Task
	if err := s.Repo.ImmediateTaskForWorker(ctx, &task); err != nil {
		log.Printf("Failed to get task: %v", err)
		return entity.Task{}, err
	}
	log.Println("s.GetNearTask: ", task)
	return task, nil
}

// очистить Zset
func (s *Service) RemoveBrokenTopTaskKey(ctx context.Context) error {
	return s.Repo.RemoveTopZSetKey(ctx)
}

// DeleteTask - удалить задачу по id
func (s *Service) DeleteTask(ctx context.Context, id string) error {
	return s.Repo.FullDeleteTask(ctx, id)
}
