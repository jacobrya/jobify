package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeStore struct {
	counts       map[string]int64
	incrErr      error
	expireErr    error
	expireCalls  int
	lastExpireAt time.Duration
}

func newFakeStore() *fakeStore {
	return &fakeStore{counts: map[string]int64{}}
}

func (f *fakeStore) Incr(ctx context.Context, key string) (int64, error) {
	if f.incrErr != nil {
		return 0, f.incrErr
	}
	f.counts[key]++
	return f.counts[key], nil
}

func (f *fakeStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	f.expireCalls++
	f.lastExpireAt = ttl
	return f.expireErr
}

func TestRateLimit_UnderLimit(t *testing.T) {
	store := newFakeStore()
	handler := RateLimit(store, 3)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "1.2.3.4:5678"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("req %d: got %d, want 200", i, rec.Code)
		}
	}
	if store.expireCalls != 1 {
		t.Errorf("expire calls: got %d, want 1 (only on first increment)", store.expireCalls)
	}
	if store.lastExpireAt != time.Minute {
		t.Errorf("expire ttl: got %v, want %v", store.lastExpireAt, time.Minute)
	}
}

func TestRateLimit_OverLimit(t *testing.T) {
	store := newFakeStore()
	handler := RateLimit(store, 2)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "9.9.9.9:1000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("req %d under limit: got %d, want 200", i, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "9.9.9.9:1000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("over limit: got %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") != "60" {
		t.Errorf("Retry-After: got %q, want 60", rec.Header().Get("Retry-After"))
	}
}

func TestRateLimit_SeparateIPs(t *testing.T) {
	store := newFakeStore()
	handler := RateLimit(store, 1)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, addr := range []string{"1.1.1.1:1", "2.2.2.2:2"} {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: got %d, want 200", addr, rec.Code)
		}
	}
}

func TestRateLimit_FailOpenOnStoreError(t *testing.T) {
	store := newFakeStore()
	store.incrErr = errors.New("redis down")
	handler := RateLimit(store, 1)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "1.2.3.4:1"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("store error should fail open: got %d, want 200", rec.Code)
	}
}

func TestRateLimit_UsesXForwardedFor(t *testing.T) {
	store := newFakeStore()
	handler := RateLimit(store, 1)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	req.RemoteAddr = "127.0.0.1:999"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	found := false
	for k := range store.counts {
		if k != "" && contains(k, "10.0.0.1") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected key to use first X-Forwarded-For IP, got counts: %v", store.counts)
	}
}

func TestRateLimit_ZeroLimitDisabled(t *testing.T) {
	store := newFakeStore()
	handler := RateLimit(store, 0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "7.7.7.7:1"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("zero limit should be disabled: got %d", rec.Code)
		}
	}
	if len(store.counts) != 0 {
		t.Errorf("store should not be called when limit=0, got %v", store.counts)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
