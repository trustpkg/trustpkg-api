package main

import (
	"context"

	"github.com/trustpkg/trustpkg-api/db"
	adaptiveWorker "github.com/trustpkg/trustpkg-api/internal/adaptive-worker"
	"github.com/trustpkg/trustpkg-api/internal/npm"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db.ConnectDb()
	db.ConnectRedis(ctx)

	worker, startWorker := adaptiveWorker.New(ctx)

	worker.ScheduleJob(adaptiveWorker.Job{
		Handler: func(concurrency int) {
			npm.Pipeline(concurrency)
		},
		SkippedCycles:    0,
		ReservationRatio: 1,
	})

	startWorker()

	<-ctx.Done()
}
