package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/trustpkg/trustpkg-api/db"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
)

func main() {
	batchSize := flag.Int("batch-size", 128, "number of package documents uploaded per batch")
	checkpointPath := flag.String("checkpoint", "tmp/typesense-packages-slug-reindex.json", "path to the resumable checkpoint")
	pause := flag.Duration("pause", 0, "pause between batches, for example 250ms")
	reset := flag.Bool("reset", false, "ignore and overwrite the existing checkpoint")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db.ConnectDb()
	defer db.Pool.Close()

	controller := typesense.NewInstance(ctx)
	if err := typesense.ReindexPackageSlugs(ctx, controller, typesense.SlugReindexOptions{
		BatchSize:      *batchSize,
		CheckpointPath: *checkpointPath,
		Pause:          *pause,
		Reset:          *reset,
	}); err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			log.Printf("Typesense slug reindex stopped: %v", err)
			return
		}
		log.Fatal(err)
	}
}
