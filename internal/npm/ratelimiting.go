package npm

import (
	"context"
	"fmt"
	"time"
)

const npmRequestInterval = 2 * time.Second

type npmRequestLimiter struct {
	tokens chan struct{}
}

func newNpmRequestLimiter(interval time.Duration) *npmRequestLimiter {
	tokens := make(chan struct{}, 1)
	tokens <- struct{}{}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}
	}()

	return &npmRequestLimiter{tokens: tokens}
}

var globalNpmRequestLimiter = newNpmRequestLimiter(npmRequestInterval)

func WaitForNpmRequest(ctx context.Context) error {
	return globalNpmRequestLimiter.wait(ctx)
}

func (limiter *npmRequestLimiter) wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("Internal Rate Limiting Error: %w", ctx.Err())
	case <-limiter.tokens:
		return nil
	}
}
