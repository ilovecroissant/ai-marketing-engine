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
    return writer.WriteMessages(ctx, kafka.Message{Value: data})
}

func Close() {
    if writer != nil {
        writer.Close()
    }
}