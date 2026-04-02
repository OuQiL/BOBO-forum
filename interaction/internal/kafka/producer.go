package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

type ProducerConfig struct {
	Brokers  []string
	Topic    string
	BatchSize int
	BatchTimeout time.Duration
}

func NewProducer(config ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		BatchSize:    config.BatchSize,
		BatchTimeout: config.BatchTimeout,
		Async:        true,
		RequiredAcks: kafka.RequireOne,
	}
	
	return &Producer{
		writer: writer,
		topic:  config.Topic,
	}
}

func (p *Producer) SendMessage(ctx context.Context, msg *LikeMessage) error {
	data, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("encode message failed: %w", err)
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d:%d", msg.TargetType, msg.TargetId)),
		Value: data,
		Time:  time.Now(),
	}

	err = p.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		return fmt.Errorf("write message to kafka failed: %w", err)
	}

	logx.Debugf("Sent like message to kafka: target_type=%d, target_id=%d, user_id=%d, action=%s",
		msg.TargetType, msg.TargetId, msg.UserId, msg.Action)
	
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
