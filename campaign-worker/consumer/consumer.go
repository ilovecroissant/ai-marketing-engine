package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
}

func New() *Consumer {
	broker := os.Getenv("KAFKA_BROKER")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          "campaigns",
		GroupID:        "campaign-worker-group", // consumer group — tracks offsets
		MinBytes:       1,
		MaxBytes:       10e6,              // 10MB max per fetch
		MaxWait:        time.Second,       // wait up to 1s for messages
		StartOffset:    kafka.FirstOffset, // start from beginning if no offset stored
		CommitInterval: time.Second,       // auto-commit offsets every 1s
	})

	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "campaigns.dlq",
		Balancer: &kafka.LeastBytes{},
	}

	return &Consumer{reader: reader, dlqWriter: dlqWriter}
}

func (c *Consumer) Start(ctx context.Context, process func(ctx context.Context, msg kafka.Message) error) {
	log.Println("Campaign consumer started")

	for {
		// FetchMessage does NOT commit the offset — you control that
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Consumer shutting down")
				return
			}
			log.Printf("Fetch error: %v", err)
			continue
		}

		log.Printf("Received message: partition=%d offset=%d key=%s",
			msg.Partition, msg.Offset, string(msg.Key))

		if err := c.processWithRetry(ctx, msg, process); err != nil {
			// All retries exhausted — send to DLQ
			log.Printf("Sending to DLQ: %s", string(msg.Key))
			c.sendToDLQ(ctx, msg, err)
		}

		// Commit offset only after processing is done
		// This is the "at-least-once" guarantee — if we crash before committing,
		// the message will be redelivered on restart
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Commit error: %v", err)
		}
	}
}

func (c *Consumer) processWithRetry(
	ctx context.Context,
	msg kafka.Message,
	process func(ctx context.Context, msg kafka.Message) error,
) error {
	maxRetries := 3
	backoff := time.Second // starts at 1s, doubles each attempt: 1s → 2s → 4s

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := process(ctx, msg)
		if err == nil {
			return nil // success
		}

		log.Printf("Attempt %d/%d failed for key=%s: %v", attempt, maxRetries, string(msg.Key), err)

		if attempt < maxRetries {
			log.Printf("Retrying in %v...", backoff)
			select {
			case <-time.After(backoff):
				backoff *= 2 // exponential backoff
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("all %d retries exhausted", maxRetries)
}

func (c *Consumer) sendToDLQ(ctx context.Context, original kafka.Message, processingErr error) {
	// Wrap original message with error metadata
	dlqPayload := map[string]any{
		"original_value": string(original.Value),
		"original_key":   string(original.Key),
		"error":          processingErr.Error(),
		"failed_at":      time.Now().UTC(),
	}

	data, _ := json.Marshal(dlqPayload)

	err := c.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   original.Key,
		Value: data,
	})
	if err != nil {
		log.Printf("Failed to write to DLQ: %v", err)
	}
}

func (c *Consumer) Close() {
	c.reader.Close()
	c.dlqWriter.Close()
}
