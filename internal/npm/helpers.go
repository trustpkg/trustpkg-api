package npm

import (
	"context"

	"github.com/trustpkg/trustpkg-api/db"
)

func getUniquePackages(packages []npmChange) []npmChange {
	unique := make(map[string]npmChange, len(packages))

	for _, item := range packages {
		unique[item.ID] = item
	}

	result := make([]npmChange, 0, len(unique))

	for _, item := range unique {
		result = append(result, item)
	}

	return result
}

func savePackagesPage(ctx context.Context, changes []npmChange) (int, error) {
	unique := getUniquePackages(changes)

	var toAdd []string
	var deprecated []string

	for _, uniquePackage := range unique {
		if uniquePackage.Deleted {
			deprecated = append(deprecated, uniquePackage.ID)
		} else {
			toAdd = append(toAdd, uniquePackage.ID)
		}
	}

	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return len(unique), err
	}
	defer transaction.Rollback(ctx)

	if len(toAdd) > 0 {
		if err := dbInsertPackagesBatch(ctx, transaction, dbInsertPackagesBatchPayload{packages: toAdd}); err != nil {
			return len(unique), err
		}
	}

	if len(deprecated) > 0 {
		if err := dbInsertMarkDeprecatedBatch(ctx, transaction, dbInsertMarkDeprecatedBatchPayload{packages: deprecated}); err != nil {
			return len(unique), err
		}
	}

	return len(unique), transaction.Commit(ctx)
}
