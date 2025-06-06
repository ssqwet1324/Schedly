package service

import (
	"auth_service/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthService struct {
	repo      *repository.Repository
	JwtSecret []byte
}

func NewAuthService(repo *repository.Repository, jwt string) *AuthService {
	return &AuthService{repo: repo, JwtSecret: []byte(jwt)}
}

// MyClaims jwt
type MyClaims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

func (service *AuthService) generateJWT(login string) (string, error) {
	claims := MyClaims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth_service",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(service.JwtSecret)
	if err != nil {
		return "", fmt.Errorf("Ошибка при создании токена: %w", err)
	}
	return tokenString, nil
}

//

func (service *AuthService) Login(ctx context.Context, user *repository.UserInfoDBO) (string, error) {
	userFromDB, err := service.repo.GetUser(ctx, user.Login)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("Пользователь %s не найден", user.Login)
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(user.Password)); err != nil {
		return "", fmt.Errorf("Неверный пароль")
	}

	jwtToken, err := service.generateJWT(user.Login)
	if err != nil {
		return "", fmt.Errorf("Не удается создать jwt токен: %w", err)
	}
	return jwtToken, nil
}

func (service *AuthService) Register(ctx context.Context, user *repository.UserInfoDBO) error {
	_, err := service.repo.GetUser(ctx, user.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("Ошибка при проверке пользователя: %w", err)
		}
	} else {
		return fmt.Errorf("Пользователь с логином %s уже зарегистрирован", user.Login)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Не удалось хешировать пароль: %w", err)
	}
	err = service.repo.AddUser(ctx, &repository.UserInfoDBO{
		Login:    user.Login,
		Password: string(hashedPassword),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("Ошибка регистрации пользователя %w", err)
		}
	}
	return nil
}
