package main

import (
	"context"

	"github.com/trustpkg/trustpkg-api/db"
	adaptiveWorker "github.com/trustpkg/trustpkg-api/internal/adaptive-worker"
	"github.com/trustpkg/trustpkg-api/internal/npm"
	"github.com/trustpkg/trustpkg-api/internal/osv"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
)

const (
	npmReplicationID    = "npm-replication"
	osvReplicationID    = "osv-replication"
	typesensePipelineID = "typesense-packages"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db.ConnectDb()
	db.ConnectRedis(ctx)
	typesenseController := typesense.NewInstance(ctx)

	worker, startWorker := adaptiveWorker.New(ctx)

	worker.ScheduleJob(adaptiveWorker.Job{
		ID: npmReplicationID,
		Handler: func() {
			if err := npm.ReplicationPipeline(); err != nil {
				npm.LogReplicatorError(err)
			}
		},
		SkippedCycles: 29,
	})

	worker.ScheduleJob(adaptiveWorker.Job{
		ID: npmReplicationID,
		Handler: func() {
			if err := npm.CollectorPipeline(ctx); err != nil {
				npm.LogCollectorError(err)
			}
		},
		SkippedCycles: 4,
	})

	worker.ScheduleJob(adaptiveWorker.Job{
		ID: osvReplicationID,
		Handler: func() {
			if err := osv.ReplicateEcosystem(ctx, osv.FetchOSVEcosystemDumpPayload{
				Ecosystem: "npm",
			}); err != nil {
				osv.LogError(err)
			}
		},
		SkippedCycles: 4,
	})

	worker.ScheduleJob(adaptiveWorker.Job{
		ID: typesensePipelineID,
		Handler: func() {
			if err := typesense.Pipeline(ctx, typesenseController); err != nil {
				typesense.LogError(err)
			}
		},
		SkippedCycles: 2,
	})

	startWorker()

	<-ctx.Done()
}
