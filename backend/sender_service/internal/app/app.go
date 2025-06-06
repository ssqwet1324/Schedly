package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"sender_service/internal/config"
	"sender_service/internal/email"
	"sender_service/internal/kafka"
)

func Run() {
	server := gin.Default()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	emailService := email.NewEmailService(
		cfg.SMTPEmail,
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
	)

	notificationConsumer := kafka.NewConsumer([]string{"kafka:9092"}, "notifications", "1", emailService)

	go func() {
		if err := notificationConsumer.StartConsumer(context.Background()); err != nil {
			log.Fatalf("Ошибка запуска consumer: %v", err)
		}
	}()

	if err := server.Run(":8092"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
