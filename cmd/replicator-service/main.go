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
		Handler: func(concurrency int, controller *adaptiveWorker.JobController) {
			npm.Pipeline(concurrency, controller)

		},
		SkippedCycles:    29,
		ReservationRatio: 0,
	})

	startWorker()

	<-ctx.Done()
}
