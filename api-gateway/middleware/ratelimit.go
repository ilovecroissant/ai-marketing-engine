package middleware

import (
    "net/http"
    "sync"

    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

var (
    limiters = make(map[string]*rate.Limiter)
    mu       sync.Mutex
)

func getLimiter(ip string) *rate.Limiter {
    mu.Lock()
    defer mu.Unlock()
    if l, ok := limiters[ip]; ok {
        return l
    }
    l := rate.NewLimiter(10, 30) // 10 req/s, burst of 30
    limiters[ip] = l
    return l
}

func RateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !getLimiter(c.ClientIP()).Allow() {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
            return
        }
        c.Next()
    }
}