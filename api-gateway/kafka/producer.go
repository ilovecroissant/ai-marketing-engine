package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func Init() {
	broker := os.Getenv("KAFKA_BROKER") // localhost:9092
	writer = &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "campaigns",
		Balancer: &kafka.LeastBytes{},
	}
	log.Println("Kafka producer initialized")
}

func PublishCampaign(ctx context.Context, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	campaignID := ""
	if m, ok := payload.(map[string]any); ok {
		if id, ok := m["campaign_id"].(string); ok {
			campaignID = id
		}
	}

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(campaignID), // ← add this
		Value: data,
	})
}

func Close() {
	if writer != nil {
		writer.Close()
	}
}
