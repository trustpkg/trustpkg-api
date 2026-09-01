package typesense

import (
	"context"
	"fmt"

	"github.com/trustpkg/trustpkg-api/internal/npm"
)

const (
	pipelineBatchSize       = 128
	pipelinePhaseWatchlist  = "watchlist"
	pipelinePhaseDatabase   = "database"
	pipelineStatusRunning   = "running"
	pipelineStatusCompleted = "completed"
)

type TypesenseWriter interface {
	CreateCollection(*CollectionPayload) error
	UpsertDocuments(*UpsertPayload) error
}

func Pipeline(ctx context.Context, writer TypesenseWriter) error {
	if writer == nil {
		typesenseLog.Print("Package pipeline skipped: Typesense writer is nil")
		return fmt.Errorf("Typesense writer cannot be nil")
	}
	if err := ensurePackageCollection(writer); err != nil {
		typesenseLog.Printf("Package pipeline collection initialization failed: %v", err)
		return err
	}

	state, err := redisGetPipelineState(ctx)
	if err != nil {
		typesenseLog.Printf("Package pipeline state read failed: %v", err)
		return err
	}
	state.Cycle++
	state.Phase = normalizePipelinePhase(state.Phase)
	typesenseLog.Printf("Package pipeline cycle started: cycle=%d phase=%s watchlist_offset=%d last_package_id=%d", state.Cycle, state.Phase, state.WatchlistOffset, state.LastPackageID)
	if err := dbSavePipelineState(ctx, state, pipelineStatusRunning, ""); err != nil {
		typesenseLog.Printf("Package pipeline state update failed: cycle=%d error=%v", state.Cycle, err)
		return err
	}

	var documents []packageDocument
	if state.Phase == pipelinePhaseWatchlist {
		documents, err = dbSelectWatchlistDocuments(ctx, npm.PackageWatchlist(), state.WatchlistOffset, pipelineBatchSize)
	} else {
		documents, err = dbSelectDatabaseDocuments(ctx, state.LastPackageID, pipelineBatchSize)
	}
	if err != nil {
		_ = dbSavePipelineState(ctx, state, "failed", err.Error())
		typesenseLog.Printf("Package pipeline document selection failed: cycle=%d phase=%s error=%v", state.Cycle, state.Phase, err)
		return err
	}

	if len(documents) == 0 && state.Phase == pipelinePhaseWatchlist {
		state.Phase = pipelinePhaseDatabase
		state.WatchlistOffset = 0
		typesenseLog.Printf("Package pipeline watchlist completed: cycle=%d switching_to=%s", state.Cycle, state.Phase)
		documents, err = dbSelectDatabaseDocuments(ctx, state.LastPackageID, pipelineBatchSize)
		if err != nil {
			_ = dbSavePipelineState(ctx, state, "failed", err.Error())
			typesenseLog.Printf("Package pipeline database document selection failed: cycle=%d error=%v", state.Cycle, err)
			return err
		}
	}
	if len(documents) == 0 {
		typesenseLog.Printf("Package pipeline completed: cycle=%d phase=%s documents_total=%d", state.Cycle, state.Phase, state.DocumentsUploaded)
		return dbSavePipelineState(ctx, state, pipelineStatusCompleted, "")
	}
	typesenseLog.Printf("Package pipeline uploading documents: cycle=%d phase=%s batch_size=%d", state.Cycle, state.Phase, len(documents))

	payload := make([]any, len(documents))
	for index := range documents {
		payload[index] = documents[index]
	}
	if err := writer.UpsertDocuments(&UpsertPayload{Name: PackageSchema.Name, BatchSize: pipelineBatchSize, Documents: payload}); err != nil {
		_ = dbSavePipelineState(ctx, state, "failed", err.Error())
		typesenseLog.Printf("Package pipeline upload failed: cycle=%d phase=%s batch_size=%d error=%v", state.Cycle, state.Phase, len(documents), err)
		return err
	}
	packageIDs := make([]int64, len(documents))
	for index := range documents {
		packageIDs[index] = packageID(documents[index].ID)
	}
	if err := dbClearPackagesNeedUpdate(ctx, packageIDs); err != nil {
		_ = dbSavePipelineState(ctx, state, "failed", err.Error())
		typesenseLog.Printf("Package pipeline update flag reset failed: cycle=%d phase=%s batch_size=%d error=%v", state.Cycle, state.Phase, len(documents), err)
		return err
	}

	if state.Phase == pipelinePhaseWatchlist {
		watchlist := npm.PackageWatchlist()
		state.WatchlistOffset = watchlistPosition(watchlist, documents[len(documents)-1].Name)
		state.WatchlistDocumentsUploaded += int64(len(documents))
		if state.WatchlistOffset >= len(watchlist) {
			state.Phase = pipelinePhaseDatabase
		}
	} else {
		state.LastPackageID = max(state.LastPackageID, packageID(documents[len(documents)-1].ID))
		state.DatabaseDocumentsUploaded += int64(len(documents))
	}
	state.DocumentsUploaded += int64(len(documents))
	if err := dbSavePipelineState(ctx, state, pipelineStatusRunning, ""); err != nil {
		typesenseLog.Printf("Package pipeline progress update failed: cycle=%d error=%v", state.Cycle, err)
		return err
	}
	typesenseLog.Printf("Package pipeline cycle completed: cycle=%d phase=%s uploaded=%d documents_total=%d", state.Cycle, state.Phase, len(documents), state.DocumentsUploaded)
	return nil
}

func ensurePackageCollection(writer TypesenseWriter) error {
	if err := writer.CreateCollection(&CollectionPayload{Name: PackageSchema.Name, Collection: PackageSchema}); err != nil && !isCollectionAlreadyExistsError(err) {
		return fmt.Errorf("initialize Typesense packages collection: %w", err)
	}
	typesenseLog.Print("Typesense packages collection is ready")
	return nil
}
