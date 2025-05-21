package sevice

import (
	"context"
	"github.com/google/uuid"
	"notif_service/internal/entity"
	"notif_service/internal/repository"
	"time"
)

type Service struct {
	Repo repository.RedisRepository
}

func NewService(repo repository.RedisRepository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateTask(ctx context.Context, e entity.Task) (entity.Task, error) {
	id := uuid.New().String()
	task := &entity.Task{
		ID:        id,
		Title:     e.Title,
		EventTime: e.EventTime,
		Body:      e.Body,
		Email:     e.Email,
	}
	err := s.Repo.SaveTask(ctx, task)
}
