package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ilovecroissant/ai-marketing-engine/campaign-worker/crawler"
	"github.com/ilovecroissant/ai-marketing-engine/campaign-worker/ranker"
	"github.com/segmentio/kafka-go"
)

type CampaignJob struct {
	CampaignID string `json:"campaign_id"`
	Industry   string `json:"industry"`
	Product    string `json:"product"`
	Website    string `json:"website"`
}

// Competitor URLs per industry — in production this would come from a DB or config
var industryURLs = map[string][]string{
	"tech": {
		"https://techcrunch.com",
		"https://www.wired.com",
		"https://arstechnica.com",
	},
	"finance": {
		"https://www.bloomberg.com",
		"https://www.ft.com",
	},
	"health": {
		"https://www.healthline.com",
		"https://www.webmd.com",
	},
}

func HandleCampaign(ctx context.Context, msg kafka.Message) error {
	var job CampaignJob
	if err := json.Unmarshal(msg.Value, &job); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	log.Printf("Processing campaign: id=%s product=%s industry=%s",
		job.CampaignID, job.Product, job.Industry)

	// Get URLs to crawl for this industry
	urls := industryURLs[strings.ToLower(job.Industry)]
	if job.Website != "" {
		urls = append(urls, job.Website) // also crawl their own site
	}
	if len(urls) == 0 {
		log.Printf("No URLs configured for industry: %s", job.Industry)
		return nil
	}

	// Run the crawler
	c := crawler.New()
	results := c.Run(ctx, urls)

	// Log results
	succeeded := 0
	for _, r := range results {
		if r.Error != nil {
			log.Printf("Crawl failed for %s: %v", r.URL, r.Error)
			continue
		}
		succeeded++
		log.Printf("Crawled %s — title=%q keywords=%v bodyLen=%d",
			r.URL, r.Title, r.Keywords, len(r.Body))
	}

	log.Printf("Campaign %s: crawled %d/%d URLs successfully",
		job.CampaignID, succeeded, len(urls))

	// Phase 6: rank and deduplicate crawled content
	rankerURL := os.Getenv("RANKER_URL")
	if rankerURL == "" {
		rankerURL = "http://localhost:8082"
	}
	rc := ranker.New(rankerURL)

	for _, r := range results {
		if r.Error != nil {
			continue
		}

		// Extract top keywords from the crawled page
		kws, err := rc.ExtractKeywords(r.Body, 10)
		if err != nil {
			log.Printf("keyword extraction failed for %s: %v", r.URL, err)
			continue
		}
		kwTerms := make([]string, len(kws))
		for i, k := range kws {
			kwTerms[i] = k.Term
		}

		// Score the content against extracted keywords
		score, err := rc.RankContent(r.Body, kwTerms)
		if err != nil {
			log.Printf("ranking failed for %s: %v", r.URL, err)
			continue
		}
		log.Printf("Ranked %s — total=%.2f seo=%.2f engagement=%.2f readability=%.2f keywords=%v",
			r.URL, score.Total, score.SEO, score.Engagement, score.Readability, kwTerms[:min(5, len(kwTerms))])
	}

	// TODO Phase 7: send ranked results to content generation
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
