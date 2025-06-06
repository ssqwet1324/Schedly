package app

import (
	"auth_service/internal/config"
	"auth_service/internal/handler"
	"auth_service/internal/migrations"
	"auth_service/internal/pkg"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"context"
	"github.com/gin-gonic/gin"
	"log"
)

func Run() {
	server := gin.New()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	server.Use(gin.Recovery(), pkg.Logger())

	repo := repository.NewRepository(cfg)
	authService := service.NewAuthService(repo, cfg.JwtSecret)
	authHandler := handler.NewAuthHandler(authService)
	migration := migrations.NewMigration(repo)

	ctx := context.Background()
	err = migration.InitTables(ctx)
	if err != nil {
		log.Fatal("Ошибка создания БД", err)
	}

	server.POST("/login", authHandler.Login)
	server.POST("/register", authHandler.Register)

	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
