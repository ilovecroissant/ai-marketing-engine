package ranker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

type Keyword struct {
	Term  string  `json:"term"`
	Score float64 `json:"score"`
}

type RankResult struct {
	SEO         float64 `json:"seo"`
	Engagement  float64 `json:"engagement"`
	Readability float64 `json:"readability"`
	Total       float64 `json:"total"`
}

type SimilarityResult struct {
	Similarity  float64 `json:"similarity"`
	IsDuplicate bool    `json:"is_duplicate"`
}

func (c *Client) ExtractKeywords(text string, n int) ([]Keyword, error) {
	body, _ := json.Marshal(map[string]any{"text": text, "n": n})
	resp, err := c.http.Post(c.baseURL+"/keywords", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ranker /keywords: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Keywords []Keyword `json:"keywords"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ranker /keywords decode: %w", err)
	}
	return result.Keywords, nil
}

func (c *Client) RankContent(text string, keywords []string) (*RankResult, error) {
	body, _ := json.Marshal(map[string]any{"text": text, "keywords": keywords})
	resp, err := c.http.Post(c.baseURL+"/rank", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ranker /rank: %w", err)
	}
	defer resp.Body.Close()

	var result RankResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ranker /rank decode: %w", err)
	}
	return &result, nil
}

func (c *Client) CheckSimilarity(a, b string) (*SimilarityResult, error) {
	body, _ := json.Marshal(map[string]string{"a": a, "b": b})
	resp, err := c.http.Post(c.baseURL+"/similarity", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ranker /similarity: %w", err)
	}
	defer resp.Body.Close()

	var result SimilarityResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ranker /similarity decode: %w", err)
	}
	return &result, nil
}
