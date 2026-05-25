package processor

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/segmentio/kafka-go"
)

type CampaignJob struct {
    CampaignID string `json:"campaign_id"`
    Industry   string `json:"industry"`
    Product    string `json:"product"`
}

func HandleCampaign(ctx context.Context, msg kafka.Message) error {
    var job CampaignJob
    if err := json.Unmarshal(msg.Value, &job); err != nil {
        // Bad JSON — retrying won't fix this, but DLQ will catch it
        return fmt.Errorf("invalid message format: %w", err)
    }

    log.Printf("Processing campaign: id=%s product=%s industry=%s",
        job.CampaignID, job.Product, job.Industry)

    // TODO Phase 5: trigger crawler with job.CampaignID
    // TODO Phase 7: trigger content generation

    log.Printf("Campaign %s processed successfully", job.CampaignID)
    return nil
}