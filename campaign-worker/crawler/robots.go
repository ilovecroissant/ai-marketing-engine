package crawler

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// isAllowed checks robots.txt before crawling a URL
// Returns true if crawling is permitted
func isAllowed(targetURL string) bool {
    parsed, err := url.Parse(targetURL)
    if err != nil {
        return false
    }

    robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsed.Scheme, parsed.Host)

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(robotsURL)
    if err != nil || resp.StatusCode != http.StatusOK {
        // If robots.txt doesn't exist or is unreachable, allow crawling
        return true
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return true
    }

    return !isDisallowed(string(body), parsed.Path)
}

func isDisallowed(robotsTxt, path string) bool {
    lines := strings.Split(robotsTxt, "\n")
    applies := false

    for _, line := range lines {
        line = strings.TrimSpace(line)

        if strings.HasPrefix(line, "User-agent:") {
            agent := strings.TrimSpace(strings.TrimPrefix(line, "User-agent:"))
            applies = agent == "*" || strings.EqualFold(agent, "AIMarketingBot")
        }

        if applies && strings.HasPrefix(line, "Disallow:") {
            disallowed := strings.TrimSpace(strings.TrimPrefix(line, "Disallow:"))
            if disallowed == "/" || strings.HasPrefix(path, disallowed) {
                return true
            }
        }
    }
    return false
}