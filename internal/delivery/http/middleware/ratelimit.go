package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/apierr"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter: per-IP token bucket (in-memory; swap for Redis when multi-instance).
type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitor
	rps        rate.Limit
	burst      int
	retryAfter time.Duration // how long one token takes to refill
}

func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitor),
		rps:        rate.Limit(rps),
		burst:      burst,
		retryAfter: refillInterval(rps),
	}
	go rl.cleanup()
	return rl
}

// Middleware: the request-limiting handler.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.limiterFor(c.ClientIP()).Allow() {
			apierr.Write(c, domain.RateLimited(domain.CodeTooManyRequests,
				"too many requests, please try again later").WithRetryAfter(rl.retryAfter))
			return
		}
		c.Next()
	}
}

func (rl *RateLimiter) limiterFor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanup: drop IPs that have been idle for a while.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// refillInterval: the wait before the bucket holds a token again, floored at one second.
func refillInterval(rps int) time.Duration {
	if rps <= 0 {
		return time.Second
	}
	if d := time.Second / time.Duration(rps); d > time.Second {
		return d
	}
	return time.Second
}
