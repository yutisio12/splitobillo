package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"splitobillo/internal/config"
	"splitobillo/pkg/apperr"
	"splitobillo/pkg/response"
)

type clientBucket struct {
	count int
	reset time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	buckets map[string]*clientBucket
}

func newRateLimiter(limit int) *rateLimiter {
	return &rateLimiter{
		limit:   limit,
		buckets: make(map[string]*clientBucket),
	}
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b := r.buckets[key]
	if b == nil || now.After(b.reset) {
		b = &clientBucket{count: 0, reset: now.Add(time.Minute)}
		r.buckets[key] = b
	}
	b.count++
	return b.count <= r.limit
}

func RateLimit(cfg *config.Config) gin.HandlerFunc {
	rl := newRateLimiter(cfg.RateLimitPerMin)
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			response.Error(c, apperr.TooManyRequests("too many requests, try again later"))
			c.Abort()
			return
		}
		c.Next()
	}
}
