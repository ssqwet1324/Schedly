package migrations

import (
	"auth_service/internal/repository"
	"context"
	"fmt"
	"log"
	"time"
)

type Migration struct {
	repo *repository.Repository
}

func NewMigration(repo *repository.Repository) *Migration {
	return &Migration{repo: repo}
}

func (m *Migration) InitTables(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS Users(
    login VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL);`

	maxRetries := 5
	retryDelay := 5 * time.Second

	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = m.repo.DB.Exec(ctx, query)
		if err == nil {
			return nil
		}

		log.Printf("Ошибка создания таблицы (попытка %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("не удалось создать таблицу после %d попыток: %w", maxRetries, err)
}
