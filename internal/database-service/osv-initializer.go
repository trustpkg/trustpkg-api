package databaseService

import (
	"context"

	"github.com/trustpkg/trustpkg-api/internal/osv"
)

type osvInitializer struct {
	ecosystem string
}

func NewOSVInitializer(ecosystem string) Initializer {
	return &osvInitializer{ecosystem: ecosystem}
}

func (initializer *osvInitializer) Needed(ctx context.Context) (bool, error) {
	initialized, err := osv.IsEcosystemInitialized(ctx, initializer.ecosystem)
	if err != nil {
		return false, err
	}
	return !initialized, nil
}

func (initializer *osvInitializer) Initialize(ctx context.Context) error {
	return osv.InitializeEcosystem(ctx, osv.FetchOSVEcosystemDumpPayload{Ecosystem: initializer.ecosystem})
}
