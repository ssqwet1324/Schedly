package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"time"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}),
	}
}

// WriteMessages - создание сообщения в кафке
func (p *Producer) WriteMessages(ctx context.Context, key string, value []byte) error {
	message := kafka.Message{
		Key:   []byte(key),
		Value: value,
	}

	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)

	defer cancelFunc()

	err := p.writer.WriteMessages(ctx, message)

	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
