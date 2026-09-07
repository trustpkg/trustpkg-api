package packagesApi

import (
	"testing"
	"time"
)

func TestNewIPRateLimiterUsesDefaults(t *testing.T) {
	limiter := NewIPRateLimiter(RateLimitConfig{Key: "packages-search"})

	if limiter.requestsPerWindow != DefaultRequestsPerWindow {
		t.Fatalf("expected %d requests per window, got %d", DefaultRequestsPerWindow, limiter.requestsPerWindow)
	}
	if limiter.requestWindow != DefaultRequestWindow {
		t.Fatalf("expected %s request window, got %s", DefaultRequestWindow, limiter.requestWindow)
	}
	if limiter.key != "packages-search" {
		t.Fatalf("expected packages-search key, got %q", limiter.key)
	}
}

func TestNewIPRateLimiterUsesCustomConfiguration(t *testing.T) {
	limiter := NewIPRateLimiter(RateLimitConfig{
		Key:               "packages-search",
		RequestsPerWindow: 25,
		RequestWindow:     30 * time.Second,
	})

	if limiter.requestsPerWindow != 25 {
		t.Fatalf("expected 25 requests per window, got %d", limiter.requestsPerWindow)
	}
	if limiter.requestWindow != 30*time.Second {
		t.Fatalf("expected 30s request window, got %s", limiter.requestWindow)
	}
	if limiter.key != "packages-search" {
		t.Fatalf("expected packages-search key, got %q", limiter.key)
	}
}

func TestRateLimitKeySeparatesEndpoints(t *testing.T) {
	searchKey := rateLimitKey("search", "127.0.0.1", 42)
	packagesKey := rateLimitKey("packages", "127.0.0.1", 42)

	if searchKey == packagesKey {
		t.Fatalf("expected endpoint keys to be different, got %q", searchKey)
	}
}
