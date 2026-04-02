package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

type Consumer struct {
	reader     *kafka.Reader
	topic      string
	groupId    string
	batchSize  int
	maxWait    time.Duration
}

type ConsumerConfig struct {
	Brokers   []string
	Topic     string
	GroupId   string
	BatchSize int
	MaxWait   time.Duration
}

func NewConsumer(config ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		GroupID:  config.GroupId,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	
	return &Consumer{
		reader:     reader,
		topic:      config.Topic,
		groupId:    config.GroupId,
		batchSize:  config.BatchSize,
		maxWait:    config.MaxWait,
	}
}

type MessageHandler func(ctx context.Context, msg *LikeMessage) error

func (c *Consumer) Start(ctx context.Context, handler MessageHandler) {
	logx.Infof("Starting kafka consumer: topic=%s, group=%s", c.topic, c.groupId)
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				logx.Info("Consumer context done, stopping...")
				return
			default:
				messages, err := c.reader.FetchMessage(ctx)
				if err != nil {
					logx.Errorf("Fetch message failed: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				
				msg, err := DecodeLikeMessage(messages.Value)
				if err != nil {
					logx.Errorf("Decode message failed: %v", err)
					c.reader.CommitMessages(ctx, messages)
					continue
				}
				
				if err := handler(ctx, msg); err != nil {
					logx.Errorf("Handle message failed: %v", err)
					continue
				}
				
				if err := c.reader.CommitMessages(ctx, messages); err != nil {
					logx.Errorf("Commit message failed: %v", err)
				}
			}
		}
	}()
}

func (c *Consumer) StartBatchConsumer(ctx context.Context, batchSize int, batchTimeout time.Duration, handler func(ctx context.Context, messages []*LikeMessage) error) {
	logx.Infof("Starting kafka batch consumer: topic=%s, group=%s, batchSize=%d", c.topic, c.groupId, batchSize)
	
	go func() {
		ticker := time.NewTicker(batchTimeout)
		defer ticker.Stop()
		
		var batch []*LikeMessage
		
		for {
			select {
			case <-ctx.Done():
				logx.Info("Batch consumer context done, stopping...")
				if len(batch) > 0 {
					c.processBatch(ctx, batch, handler)
				}
				return
				
			case <-ticker.C:
				if len(batch) > 0 {
					c.processBatch(ctx, batch, handler)
					batch = batch[:0]
				}
				
			default:
				messages, err := c.reader.FetchMessage(ctx)
				if err != nil {
					logx.Debugf("Fetch message failed: %v", err)
					time.Sleep(50 * time.Millisecond)
					continue
				}
				
				msg, err := DecodeLikeMessage(messages.Value)
				if err != nil {
					logx.Errorf("Decode message failed: %v", err)
					c.reader.CommitMessages(ctx, messages)
					continue
				}
				
				batch = append(batch, msg)
				
				if len(batch) >= batchSize {
					c.processBatch(ctx, batch, handler)
					batch = batch[:0]
				}
			}
		}
	}()
}

func (c *Consumer) processBatch(ctx context.Context, messages []*LikeMessage, handler func(ctx context.Context, messages []*LikeMessage) error) {
	logx.Infof("Processing batch of %d messages", len(messages))
	
	if err := handler(ctx, messages); err != nil {
		logx.Errorf("Process batch failed: %v", err)
		return
	}
	
	logx.Infof("Successfully processed batch of %d messages", len(messages))
}

func (c *Consumer) Close() error {
	logx.Info("Closing kafka consumer...")
	return c.reader.Close()
}
