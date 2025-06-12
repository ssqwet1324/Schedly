package worker

import (
	"context"
	"encoding/json"
	"log"
	"notif_service/internal/entity"
	"notif_service/internal/kafka"
	"notif_service/internal/service"
	"time"
)

type Worker struct {
	ctx      context.Context
	service  *service.Service
	producer *kafka.Producer
	wakeup   chan bool
}

func NewWorker(ctx context.Context, service *service.Service, producer *kafka.Producer) *Worker {
	return &Worker{
		ctx:      ctx,
		service:  service,
		producer: producer,
		wakeup:   make(chan bool, 1),
	}
}

// Run - запуск воркера
func (w *Worker) Run() {
	for {
		select {
		case <-w.ctx.Done():
			log.Println("Worker exit")
			return

		case <-w.wakeup:
			log.Println("Worker wakeup")
		default:
			task, err := w.service.GetNearTask(w.ctx)
			if err != nil {
				if err.Error() == "redis: nil" {
					//ожидание сигнала
					w.waitForSignal()
					continue
				}
				log.Printf("Ошибка при получении задачи: %v", err)
				time.Sleep(time.Second)
				continue
			}
			//проверка на пустую таску
			if !w.validateTask(task) {
				continue
			}

			duration := time.Until(task.EventTime)
			if duration <= 0 {
				// проверка на прошедшую задачу
				w.handleOverdueTask(task)
				continue
			}

			log.Printf("Следующая задача [%s] через %v", task.Title, duration)
			//ожидание отправления таски
			w.waitForTaskTime(task)
		}
	}
}

// validateTask - проверяет валидность задачи
func (w *Worker) validateTask(task entity.Task) bool {
	if task.ID == "" {
		log.Println("Получена пустая задача. Ждем сигнал...")
		select {
		case <-w.wakeup:
		case <-w.ctx.Done():
		}
		return false
	}
	return true
}

// handleOverdueTask - обрабатывает просроченную задачу
func (w *Worker) handleOverdueTask(task entity.Task) {
	log.Printf("Задача [%s] уже просрочена или устарела", task.Title)
	if err := w.service.DeleteTask(w.ctx, task.ID); err != nil {
		log.Printf("Ошибка удаления просроченной задачи: %v", err)
	}
}

// waitForSignal - ожидание сигнала пробуждения
func (w *Worker) waitForSignal() {
	log.Println("Нет задач. Ожидаем сигнал...")
	select {
	case <-w.wakeup:
		log.Println("Проснулся по сигналу")
	case <-w.ctx.Done():
		log.Println("Worker exit while waiting")
	}
}

// waitForTaskTime - ожидание времени выполнения задачи
func (w *Worker) waitForTaskTime(task entity.Task) {
	timer := time.NewTimer(time.Until(task.EventTime))
	select {
	case <-w.ctx.Done():
		log.Println("Worker stopped")
		timer.Stop()
		return
	case <-timer.C:
		log.Printf("Время задачи [%s] пришло", task.Title)
		w.sendNotification(task, "Время задачи пришло")
		w.processAllDueTasks()
	case <-w.wakeup:
		log.Println("Получен сигнал до наступления времени задачи")
		timer.Stop()
	}
}

// processAllDueTasks - обрабатывает все задачи, которые должны быть отправлены сейчас
func (w *Worker) processAllDueTasks() {
	for {
		task, err := w.service.GetNearTask(w.ctx)
		if err != nil {
			if err.Error() == "redis: nil" {
				return
			}
			log.Printf("Ошибка при получении следующей задачи: %v", err)
			return
		}

		if !w.validateTask(task) {
			return
		}

		if time.Until(task.EventTime) <= 0 {
			log.Printf("Отправка следующей задачи [%s]", task.Title)
			w.sendNotification(task, "Время задачи пришло")
		} else {
			return
		}
	}
}

// sendNotification - отправка уведомления в кафку
func (w *Worker) sendNotification(task entity.Task, message string) {
	notify := map[string]interface{}{
		"title":      task.Title,
		"event_time": task.EventTime,
		"message":    message,
		"email":      task.Email,
		"body":       task.Body,
	}

	if err := w.service.DeleteTask(w.ctx, task.ID); err != nil {
		log.Printf("Ошибка при удалении задачи [%s]: %v", task.Title, err)
	}

	value, err := json.Marshal(notify)
	if err != nil {
		log.Printf("Ошибка маршала уведомления [%s]: %v", task.Title, err)
		return
	}

	if err := w.producer.WriteMessages(w.ctx, task.Title, value); err != nil {
		log.Printf("Ошибка отправки уведомления в Kafka [%s]: %v", task.Title, err)
	}
}

// RedefinitionWorker - переопределение воркера при дерганее ручки
func (w *Worker) RedefinitionWorker() {
	select {
	case w.wakeup <- true:
		log.Println("Сигнал на пробуждение воркера отправлен")
	default:
		log.Println("Сигнал на пробуждение воркера уже отправлен")
	}
}
