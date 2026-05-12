package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers      []string      `mapstructure:"Brokers"`
	Topic        string        `mapstructure:"Topic"`
	ClientID     string        `mapstructure:"ClientID"`
	DialTimeout  time.Duration `mapstructure:"DialTimeout"`
	WriteTimeout time.Duration `mapstructure:"WriteTimeout"`
}

type kafkaPublisher struct {
	writer *kafkaGo.Writer
}

type noopPublisher struct{}

func NewPublisher(cfg Config) (Publisher, error) {
	if len(cfg.Brokers) == 0 {
		return noopPublisher{}, nil
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	writer := kafkaGo.NewWriter(kafkaGo.WriterConfig{
		Brokers:      cfg.Brokers,
		Topic:        cfg.Topic,
		Balancer:     &kafkaGo.Hash{},
		Dialer:       &kafkaGo.Dialer{Timeout: cfg.DialTimeout, ClientID: cfg.ClientID},
		BatchTimeout: cfg.WriteTimeout,
	})

	return &kafkaPublisher{writer: writer}, nil
}

func (p *kafkaPublisher) Publish(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal kafka event: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafkaGo.Message{
		Key:   []byte(event.EntityID),
		Value: payload,
	})
}

func (p *kafkaPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func (noopPublisher) Publish(ctx context.Context, event Event) error {
	return nil
}

func (noopPublisher) Close() error {
	return nil
}
