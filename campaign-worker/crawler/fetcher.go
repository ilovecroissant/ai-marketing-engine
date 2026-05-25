package crawler

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "time"
	"io"
	"sort"

    "golang.org/x/net/html"
)

type CrawlResult struct {
    URL      string
    Title    string
    Body     string
    Keywords []string
    Error    error
}

func fetch(ctx context.Context, url string) CrawlResult {
    // Create request with context so it respects timeouts
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return CrawlResult{URL: url, Error: fmt.Errorf("bad request: %w", err)}
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AIMarketingBot/1.0)")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return CrawlResult{URL: url, Error: fmt.Errorf("fetch failed: %w", err)}
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return CrawlResult{URL: url, Error: fmt.Errorf("status %d", resp.StatusCode)}
    }

    title, body, err := parseHTML(resp.Body)
    if err != nil {
        return CrawlResult{URL: url, Error: fmt.Errorf("parse failed: %w", err)}
    }

    return CrawlResult{
        URL:      url,
        Title:    title,
        Body:     body,
        Keywords: extractKeywords(body),
    }
}

func parseHTML(r io.Reader) (title, body string, err error) {
    doc, err := html.Parse(r)
    if err != nil {
        return "", "", err
    }

    var sb strings.Builder

    var walk func(*html.Node)
    walk = func(n *html.Node) {
        if n.Type == html.ElementNode {
            // Skip scripts, styles, nav — we only want readable content
            switch n.Data {
            case "script", "style", "nav", "footer", "header":
                return
            case "title":
                if n.FirstChild != nil {
                    title = n.FirstChild.Data
                }
            }
        }
        if n.Type == html.TextNode {
            text := strings.TrimSpace(n.Data)
            if len(text) > 20 { // ignore short fragments like ">" or whitespace
                sb.WriteString(text)
                sb.WriteString(" ")
            }
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            walk(c)
        }
    }
    walk(doc)

    return title, sb.String(), nil
}

func extractKeywords(body string) []string {
    // Simple frequency-based extraction — Phase 6 C++ engine will do TF-IDF properly
    words := strings.Fields(strings.ToLower(body))
    freq := make(map[string]int)

    stopWords := map[string]bool{
        "the": true, "and": true, "for": true, "are": true,
        "but": true, "not": true, "you": true, "all": true,
        "can": true, "her": true, "was": true, "one": true,
        "our": true, "out": true, "has": true, "have": true,
    }

    for _, w := range words {
        if len(w) > 3 && !stopWords[w] {
            freq[w]++
        }
    }

    // Pick top 10 by frequency
    type kv struct {
        word  string
        count int
    }
    var pairs []kv
    for w, c := range freq {
        pairs = append(pairs, kv{w, c})
    }
    sort.Slice(pairs, func(i, j int) bool {
        return pairs[i].count > pairs[j].count
    })

    keywords := make([]string, 0, 10)
    for i := 0; i < 10 && i < len(pairs); i++ {
        keywords = append(keywords, pairs[i].word)
    }
    return keywords
}