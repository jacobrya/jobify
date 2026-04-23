package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abzalserikbay/jobify/pkg/response"
)

type RateLimitStore interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

func RateLimit(store RateLimitStore, limitPerMin int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limitPerMin <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			ip := clientIP(r)
			minute := time.Now().Unix() / 60
			key := "rl:" + ip + ":" + strconv.FormatInt(minute, 10)

			count, err := store.Incr(r.Context(), key)
			if err != nil {
				// fail open: a broken limiter must not take down the API
				next.ServeHTTP(w, r)
				return
			}
			if count == 1 {
				_ = store.Expire(r.Context(), key, time.Minute)
			}
			if int(count) > limitPerMin {
				w.Header().Set("Retry-After", "60")
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		return strings.TrimSpace(first)
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx >= 0 {
		host = host[:idx]
	}
	return host
}
