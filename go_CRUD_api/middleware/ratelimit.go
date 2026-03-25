package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/codemarked/go-lab/go_CRUD_api/api"
	"github.com/codemarked/go-lab/go_CRUD_api/respond"
	"github.com/gin-gonic/gin"
)

type ipWindow struct {
	windowStart time.Time
	count       int
}

// NewFixedWindowLimiter limits requests per client IP in a fixed 1-minute window.
// In-memory only: not shared across replicas (see docs/auth-session.md abuse section).
func NewFixedWindowLimiter(rpm int, message string) gin.HandlerFunc {
	if rpm <= 0 {
		rpm = 30
	}
	if message == "" {
		message = "too many requests"
	}
	var mu sync.Mutex
	byIP := make(map[string]*ipWindow)
	const win = time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		mu.Lock()
		w, ok := byIP[ip]
		if !ok {
			w = &ipWindow{windowStart: now, count: 0}
			byIP[ip] = w
		}
		if now.Sub(w.windowStart) >= win {
			w.windowStart = now
			w.count = 0
		}
		if w.count >= rpm {
			mu.Unlock()
			respond.Error(c, http.StatusTooManyRequests, api.CodeRateLimited, message, nil)
			c.Abort()
			return
		}
		w.count++
		mu.Unlock()
		c.Next()
	}
}

// NewTokenEndpointLimiter limits token endpoint calls per client IP (fixed 1-minute window).
func NewTokenEndpointLimiter(rpm int) gin.HandlerFunc {
	return NewFixedWindowLimiter(rpm, "too many token requests")
}
