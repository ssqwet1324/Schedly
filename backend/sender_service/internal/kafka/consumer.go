package kafka

import (
	"context"
	"encoding/json"
	"fmt"
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
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Println("Ошибка закрытия Kafka reader:", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped")
			return nil
		default:
			msg, err := c.readMessage(ctx)
			if err != nil {
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Ошибка обработки сообщения: %v", err)
			}
		}
	}
}

// readMessage() - чтение сообщение в кафке
func (c *Consumer) readMessage(ctx context.Context) (kafka.Message, error) {
	readCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return c.reader.ReadMessage(readCtx)
}

// processMessage() - отправка сообщения на почту
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var data entity.EmailMessage
	if err := json.Unmarshal(msg.Value, &data); err != nil {
		return fmt.Errorf("ошибка парсинга: %w", err)
	}

	if err := c.emailService.SendEmail(data); err != nil {
		return fmt.Errorf("ошибка отправки email: %w", err)
	}

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("ошибка коммита: %w", err)
	}

	return nil
}
