package npm

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"
)

const testNpmRequestInterval = 20 * time.Millisecond

func newTestNpmRequestLimiter() *npmRequestLimiter {
	return newNpmRequestLimiter(testNpmRequestInterval)
}

func TestNpmRequestIntervalIsTwoSeconds(t *testing.T) {
	if npmRequestInterval != 2*time.Second {
		t.Fatalf("npm request interval is %v; expected 2s", npmRequestInterval)
	}
}

func TestNpmRequestLimiterAllowsFirstRequestImmediately(t *testing.T) {
	limiter := newTestNpmRequestLimiter()
	start := time.Now()

	if err := limiter.wait(context.Background()); err != nil {
		t.Fatalf("wait() returned an unexpected error: %v", err)
	}

	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("first request waited %v; expected it to be available immediately", elapsed)
	}
}

func TestNpmRequestLimiterEnforcesTwoSecondInterval(t *testing.T) {
	limiter := newTestNpmRequestLimiter()

	if err := limiter.wait(context.Background()); err != nil {
		t.Fatalf("wait() for the first request returned an unexpected error: %v", err)
	}

	start := time.Now()
	if err := limiter.wait(context.Background()); err != nil {
		t.Fatalf("wait() for the second request returned an unexpected error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 15*time.Millisecond {
		t.Fatalf("second request was released after %v; expected at least 15ms", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Fatalf("second request was released after %v; expected less than 100ms", elapsed)
	}
}

func TestNpmRequestLimiterQueuesConcurrentRequests(t *testing.T) {
	limiter := newTestNpmRequestLimiter()
	const requestCount = 3

	start := time.Now()
	ready := make(chan struct{})
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(requestCount)
	elapsed := make([]time.Duration, 0, requestCount)
	var elapsedMutex sync.Mutex

	for range requestCount {
		go func() {
			defer waitGroup.Done()
			<-ready

			if err := limiter.wait(context.Background()); err != nil {
				t.Errorf("wait() returned an unexpected error: %v", err)
				return
			}

			elapsedMutex.Lock()
			elapsed = append(elapsed, time.Since(start))
			elapsedMutex.Unlock()
		}()
	}

	close(ready)
	waitGroup.Wait()
	sort.Slice(elapsed, func(first, second int) bool {
		return elapsed[first] < elapsed[second]
	})

	if len(elapsed) != requestCount {
		t.Fatalf("got %d completed requests; expected %d", len(elapsed), requestCount)
	}
	if elapsed[0] > 100*time.Millisecond {
		t.Fatalf("first concurrent request completed after %v", elapsed[0])
	}
	if elapsed[1] < 15*time.Millisecond {
		t.Fatalf("second concurrent request completed after %v; expected at least 15ms", elapsed[1])
	}
	if elapsed[2] < 35*time.Millisecond {
		t.Fatalf("third concurrent request completed after %v; expected at least 35ms", elapsed[2])
	}
}

func TestNpmRequestLimiterDoesNotAccumulateMoreThanOneWaitingToken(t *testing.T) {
	limiter := newTestNpmRequestLimiter()

	if err := limiter.wait(context.Background()); err != nil {
		t.Fatalf("wait() for the first request returned an unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := limiter.wait(context.Background()); err != nil {
		t.Fatalf("wait() for the second request returned an unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	if err := limiter.wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("third request returned %v; expected no accumulated token", err)
	}
}

func TestNpmRequestLimiterHonorsContextCancellation(t *testing.T) {
	limiter := newTestNpmRequestLimiter()
	if err := limiter.wait(context.Background()); err != nil {
		t.Fatalf("wait() for the first request returned an unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err := limiter.wait(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wait() returned %v; expected context deadline exceeded", err)
	}
}
