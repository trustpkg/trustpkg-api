package osv

import (
	"context"
	"fmt"
	"strings"
)

func IsEcosystemInitialized(ctx context.Context, ecosystem string) (bool, error) {
	if strings.TrimSpace(ecosystem) == "" {
		return false, fmt.Errorf("OSV ecosystem cannot be empty")
	}

	state, err := dbGetOsvReplicationState(ctx, ecosystem)
	if err != nil {
		return false, fmt.Errorf("get OSV replication state: %w", err)
	}
	return state.IsFirstSynchronised, nil
}

func InitializeEcosystem(ctx context.Context, payload FetchOSVEcosystemDumpPayload) error {
	return newReplicator(ctx, payload).RunInitial()
}
