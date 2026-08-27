package osv

import (
	"context"
	"fmt"

	"github.com/trustpkg/trustpkg-api/db"
)

func SyncPackage(ctx context.Context, payload SyncPackagePayload) error {
	data, err := fetchPackageVulnerabilities(ctx, fetchPackageVulnerabilitiesPayload{
		name:      payload.Name,
		ecosystem: payload.Ecosystem,
	})
	if err != nil {
		return fmt.Errorf("Fetching package vulnerability error: %w", err)
	}

	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Starting package synchronization transaction error: %w", err)
	}
	defer transaction.Rollback(ctx)

	if err := deletePackageVulnerabilities(ctx, transaction, payload.Name, payload.Ecosystem); err != nil {
		return fmt.Errorf("Removing old package vulnerabilities error: %w", err)
	}

	for _, vulnerability := range data.Vulns {
		if err := saveVulnerabilityTx(ctx, transaction, parseVulnerability(vulnerability)); err != nil {
			return fmt.Errorf("Saving package vulnerability error: %w", err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("Committing package synchronization error: %w", err)
	}

	return nil
}
