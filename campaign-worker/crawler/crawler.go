package crawler

import (
    "context"
    "log"
    "sync"
    "time"
	"fmt"
)

const (
    NumWorkers    = 10              // max concurrent fetches
    FetchTimeout  = 15 * time.Second
)

type Crawler struct {
    numWorkers int
}

func New() *Crawler {
    return &Crawler{numWorkers: NumWorkers}
}

func (c *Crawler) Run(ctx context.Context, urls []string) []CrawlResult {
    // Channel of URLs to process — acts as the work queue
    urlCh := make(chan string, len(urls))

    // Channel of results — workers write here
    resultCh := make(chan CrawlResult, len(urls))

    // Feed all URLs into the channel upfront and close it
    // Workers will drain it until empty
    for _, u := range urls {
        urlCh <- u
    }
    close(urlCh) // closing tells workers "no more work coming"

    // Launch exactly N workers — no more will ever run simultaneously
    var wg sync.WaitGroup
    for i := 0; i < c.numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            worker(ctx, workerID, urlCh, resultCh)
        }(i)
    }

    // Wait for all workers to finish, then close results channel
    // This runs in a goroutine so we don't block collecting results
    go func() {
        wg.Wait()
        close(resultCh)
    }()

    // Collect all results
    var results []CrawlResult
    for r := range resultCh {
        results = append(results, r)
    }

    return results
}

func worker(ctx context.Context, id int, urls <-chan string, results chan<- CrawlResult) {
    for url := range urls {
        // Check context before each fetch — respects cancellation/timeout
        select {
        case <-ctx.Done():
            log.Printf("Worker %d: context cancelled, stopping", id)
            return
        default:
        }

        // Check robots.txt
        if !isAllowed(url) {
            log.Printf("Worker %d: robots.txt disallows %s, skipping", id, url)
            results <- CrawlResult{URL: url, Error: fmt.Errorf("disallowed by robots.txt")}
            continue
        }

        log.Printf("Worker %d: fetching %s", id, url)

        // Create per-fetch context with timeout
        fetchCtx, cancel := context.WithTimeout(ctx, FetchTimeout)
        result := fetchWithRetry(fetchCtx, url)
        cancel()

        results <- result
    }
}

func fetchWithRetry(ctx context.Context, url string) CrawlResult {
    backoff := time.Second
    maxRetries := 3

    for attempt := 1; attempt <= maxRetries; attempt++ {
        result := fetch(ctx, url)
        if result.Error == nil {
            return result
        }

        log.Printf("Fetch attempt %d/%d failed for %s: %v", attempt, maxRetries, url, result.Error)

        if attempt < maxRetries {
            select {
            case <-time.After(backoff):
                backoff *= 2 // 1s → 2s → 4s
            case <-ctx.Done():
                return CrawlResult{URL: url, Error: ctx.Err()}
            }
        }
    }

    return CrawlResult{URL: url, Error: fmt.Errorf("all retries exhausted")}
}