package repository

import (
	"auth_service/internal/config"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
)

type Repository struct {
	DB *pgxpool.Pool
}

type UserInfoDBO struct {
	Login    string
	Password string
}

func NewRepository(cfg *config.Config) *Repository {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		strconv.Itoa(cfg.DbPort),
		cfg.DbName,
	)
	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	return &Repository{dbpool}
}

func (repo *Repository) GetUser(ctx context.Context, login string) (*UserInfoDBO, error) {
	var user UserInfoDBO

	err := repo.DB.QueryRow(ctx, "SELECT login, password FROM Users WHERE login = $1", login).
		Scan(&user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &UserInfoDBO{}, fmt.Errorf("Пользователь с логином %s не найден", login)
		}
		return &UserInfoDBO{}, fmt.Errorf("Ошибка получения информации о пользователе: %w", err)
	}
	return &user, nil
}

func (repo *Repository) AddUser(ctx context.Context, user *UserInfoDBO) error {

	query := "INSERT INTO Users (login, password) VALUES ($1, $2)"

	_, err := repo.DB.Exec(ctx, query, user.Login, user.Password)
	if err != nil {
		return fmt.Errorf("Ошибка добавления пользователя: %w", err)
	}
	return nil
}
