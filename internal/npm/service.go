package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/trustpkg/trustpkg-api/db"
	adaptiveWorker "github.com/trustpkg/trustpkg-api/internal/adaptive-worker"
)

func Pipeline(concurrency int, JobController *adaptiveWorker.JobController) error {
	ctx := context.Background()

	if concurrency < 1 {
		concurrency = 1
	}

	since, err := selectLastSequence(ctx)
	if err != nil {
		return err
	}

	var waitGroup sync.WaitGroup

	errs := make(chan error, concurrency)

	for page := 0; page < concurrency; page++ {
		response, err := fetchNpmPackages(since)
		if err != nil {
			return err
		}

		since = response.LastSeq

		waitGroup.Add(1)

		go func(changes []npmChange) {
			defer waitGroup.Done()

			if err := savePackagesPage(ctx, changes); err != nil {
				errs <- err
			}

			fmt.Print("---------------------->\n")
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(response.Results)

			fmt.Println("response", response.LastSeq)
			fmt.Print("----------------------<\n")
		}(response.Results)
	}

	waitGroup.Wait()
	close(errs)

	data, err := FetchNpmPackageDownloads(FetchNpmPackageDownloadsPayload{
		packageName: "next",
	})

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(data)

	for err := range errs {
		if err != nil {
			return err
		}
	}

	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	if err := insertLastSequence(ctx, transaction, since); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func savePackagesPage(ctx context.Context, changes []npmChange) error {
	unique := getUniquePackages(changes)

	var toAdd []string
	var toRemove []string

	for _, uniquePackage := range unique {
		if uniquePackage.Deleted {
			toRemove = append(toRemove, uniquePackage.ID)
		} else {
			toAdd = append(toAdd, uniquePackage.ID)
		}
	}

	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	if len(toAdd) > 0 {
		if err := insertPackagesBatch(ctx, transaction, dbInsertPackagesBatchPayload{packages: toAdd}); err != nil {
			return err
		}
	}

	if len(toRemove) > 0 {
		if err := dropPackagesBatch(ctx, transaction, dbDropPackagesBatchPayload{packages: toRemove}); err != nil {
			return err
		}
	}

	return transaction.Commit(ctx)
}
