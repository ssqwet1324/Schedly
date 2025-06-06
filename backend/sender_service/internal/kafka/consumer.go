package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/segmentio/kafka-go"
	"log"
	"sender_service/internal/email"
	"sender_service/internal/entity"
	"time"
)

type Consumer struct {
	reader       *kafka.Reader
	emailService *email.EmailService
}

type Reader interface {
	SendEmail(ctx context.Context, message entity.EmailMessage) error
}

func NewConsumer(brokers []string, topic, groupID string, service *email.EmailService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumer{reader: reader, emailService: service}
}

// StartConsumer - запуск consumer
func (c *Consumer) StartConsumer(ctx context.Context) error {
	defer func(reader *kafka.Reader) {
		err := reader.Close()
		if err != nil {
			log.Println("Ошибка закрытия Kafka reader:", err)
		}
	}(c.reader)

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped")
			_ = c.reader.Close()
			return nil // завершаем корректно
		default:
			readCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			msg, err := c.reader.ReadMessage(readCtx)
			cancel()

			if err != nil {
				// Убираем флуд: логируем только при необходимости
				if !errors.Is(err, context.DeadlineExceeded) {
					log.Printf("Ошибка чтения сообщения: %v", err)
				}
				continue
			}

			log.Println("Consumer Read Message: ", string(msg.Value))

			var data entity.EmailMessage
			err = json.Unmarshal(msg.Value, &data)
			if err != nil {
				log.Printf("Ошибка распарсивания данных: %v", err)
				continue
			}

			err = c.emailService.SendEmail(data)
			if err != nil {
				log.Printf("Ошибка отправки email: %v", err)
				continue
			}

			err = c.reader.CommitMessages(ctx, msg)
			if err != nil {
				log.Println("Ошибка коммита сообщения:", err)
				continue
			}
		}
	}
}
