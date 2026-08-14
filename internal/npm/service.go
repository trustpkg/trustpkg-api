package npm

import (
	"context"
	"fmt"

	"github.com/trustpkg/trustpkg-api/db"
	adaptiveWorker "github.com/trustpkg/trustpkg-api/internal/adaptive-worker"
)

func Pipeline() error {
	ctx := context.Background()

	response, err := fetchNpmPackagesByCurrentSequence(ctx)
	if err != nil {
		return err
	}

	unique := getUniquePackages(response.Results)

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
		err := insertPackagesBatch(ctx, transaction, dbInsertPackagesBatchPayload{packages: toAdd})
		if err != nil {
			return err
		}
	}

	if len(toRemove) > 0 {
		err := dropPackagesBatch(ctx, transaction, dbDropPackagesBatchPayload{packages: toRemove})
		if err != nil {
			return err
		}
	}

	fmt.Println("overrite Sequence: ", response.LastSeq)

	err = insertLastSequence(ctx, transaction, response.LastSeq)
	if err != nil {
		return err
	}

	data, _ := adaptiveWorker.CheckResourceUsage()
	fmt.Println("data", data)

	return transaction.Commit(ctx)
}
