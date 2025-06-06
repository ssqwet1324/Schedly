package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"notif_service/internal/handler"
	"notif_service/internal/kafka"
	"notif_service/internal/repository"
	service "notif_service/internal/service"
	"notif_service/internal/worker"
)

func Run() {
	server := gin.Default()

	repo := repository.NewRedisRepository("redis:6379", "", 0)

	producer := kafka.NewProducer([]string{"kafka:9092"}, "notifications")

	notificationService := service.NewService(*repo, producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newWorker := worker.NewWorker(ctx, notificationService, producer)

	go newWorker.Run()

	notificationHandler := handler.NewHandler(notificationService, newWorker)

	server.POST("/createTask", notificationHandler.CreateTaskHandler)

	server.POST("/deleteTask", notificationHandler.DeleteTaskHandler)

	if err := server.Run(":8091"); err != nil {
		log.Fatal(err)
	}
}
