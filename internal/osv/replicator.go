package osv

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trustpkg/trustpkg-api/db"
)

const (
	maxOSVWorkers = 6
	osvWritePause = 10 * time.Millisecond
)

type Replicator struct {
	ctx     context.Context
	payload FetchOSVEcosystemDumpPayload
	state   *OsvReplicationState
}

func newReplicator(ctx context.Context, payload FetchOSVEcosystemDumpPayload) *Replicator {
	return &Replicator{ctx: ctx, payload: payload}
}

func ReplicateEcosystem(ctx context.Context, payload FetchOSVEcosystemDumpPayload) error {
	return newReplicator(ctx, payload).Run()
}

func (replicator *Replicator) Run() error {
	ctx := replicator.ctx
	payload := replicator.payload
	osvLog.Printf("OSV replication started: ecosystem=%s", payload.Ecosystem)
	if strings.TrimSpace(payload.Ecosystem) == "" {
		return fmt.Errorf("OSV ecosystem cannot be empty")
	}

	state, err := dbGetOsvReplicationState(ctx, payload.Ecosystem)
	if err != nil {
		return fmt.Errorf("get OSV replication state: %w", err)
	}
	replicator.state = state
	if state.IsFirstSynchronised {
		osvLog.Printf("OSV replication mode: incremental, ecosystem=%s, last_modified_at=%v, next_check_at=%v", payload.Ecosystem, state.LastModifiedAt, state.NextCheckAt)
		return replicator.runIncremental()
	}
	osvLog.Printf("OSV replication mode: full, ecosystem=%s", payload.Ecosystem)
	nextCheckAt := time.Now().Add(osvReplicationCooldown)
	claimed, err := dbClaimOsvReplicationCheck(ctx, payload.Ecosystem, nextCheckAt)
	if err != nil {
		return fmt.Errorf("claim OSV full replication: %w", err)
	}
	if !claimed {
		osvLog.Printf("OSV full replication skipped: ecosystem=%s, cooldown active, next check at %s, last check at %s", payload.Ecosystem, formatOsvTime(state.NextCheckAt), formatOsvTime(state.LastCheckedAt))
		return nil
	}
	osvLog.Printf("OSV cooldown started: ecosystem=%s, mode=full, reason=full dump download, next check at %s", payload.Ecosystem, formatOsvTime(&nextCheckAt))

	lastModifiedAt, err := replicator.runFull()
	if err != nil {
		osvLog.Printf("OSV cooldown retained: ecosystem=%s, mode=full, reason=full replication failed, next check at %s, error=%v", payload.Ecosystem, formatOsvTime(&nextCheckAt), err)
		return err
	}
	if err := dbCompleteOsvFullReplication(ctx, payload.Ecosystem, lastModifiedAt); err != nil {
		osvLog.Printf("OSV cooldown retained: ecosystem=%s, mode=full, reason=state update failed, next check at %s, error=%v", payload.Ecosystem, formatOsvTime(&nextCheckAt), err)
		return fmt.Errorf("complete OSV full replication: %w", err)
	}
	osvLog.Printf("OSV full replication completed: ecosystem=%s, last_modified_at=%v", payload.Ecosystem, lastModifiedAt)

	return nil
}

func (replicator *Replicator) runFull() (*time.Time, error) {
	osvLog.Printf("OSV full dump download started: ecosystem=%s", replicator.payload.Ecosystem)
	dump, err := fetchOSVEcosystemDump(replicator.ctx, replicator.payload)
	if err != nil {
		return nil, fmt.Errorf("fetch OSV dump: %w", err)
	}

	archiveData, err := io.ReadAll(dump)
	dump.Close()
	if err != nil {
		return nil, fmt.Errorf("read OSV dump: %w", err)
	}

	archive, err := zip.NewReader(byteReaderAt(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, fmt.Errorf("open OSV ZIP: %w", err)
	}
	osvLog.Printf("OSV full dump loaded: ecosystem=%s, files=%d, size_mb=%.2f", replicator.payload.Ecosystem, len(archive.File), float64(len(archiveData))/(1024*1024))

	workerContext, cancel := context.WithCancel(replicator.ctx)
	defer cancel()

	processingReporter := newStatusReporter("OSV PostgreSQL", "records", int64(len(archive.File)), 1)
	jobs := make(chan ParsedVulnerability, maxOSVWorkers)

	var workers sync.WaitGroup
	var saved atomic.Int64
	var firstErr error
	var errorOnce sync.Once
	var newestModifiedAt *time.Time

	for range maxOSVWorkers {
		workers.Go(func() {
			for vulnerability := range jobs {
				if err := dbSaveVulnerability(workerContext, vulnerability); err != nil {
					errorOnce.Do(func() {
						firstErr = err
						cancel()
					})
				} else {
					processingReporter.update(saved.Add(1))
					if !waitForOSVWritePause(workerContext) {
						return
					}
				}
			}
		})
	}

fileLoop:
	for _, file := range archive.File {
		if workerContext.Err() != nil {
			break
		}
		if file.FileInfo().IsDir() || file.Name == "" || len(file.Name) < 5 || file.Name[len(file.Name)-5:] != ".json" {
			continue
		}

		reader, openErr := file.Open()
		if openErr != nil {
			errorOnce.Do(func() { firstErr = openErr })
			break
		}

		var raw OsvVuln
		decodeErr := json.NewDecoder(reader).Decode(&raw)
		reader.Close()
		if decodeErr != nil {
			errorOnce.Do(func() {
				firstErr = fmt.Errorf("decode %s: %w", file.Name, decodeErr)
				cancel()
			})
			break
		}
		if raw.Modified.After(time.Time{}) && (newestModifiedAt == nil || raw.Modified.After(*newestModifiedAt)) {
			modifiedAt := raw.Modified
			newestModifiedAt = &modifiedAt
		}
		select {
		case jobs <- parseVulnerability(raw):
		case <-workerContext.Done():
			break fileLoop
		}
	}
	close(jobs)

	workers.Wait()
	if firstErr != nil {
		processingReporter.finish("OSV PostgreSQL stage failed")
		return nil, firstErr
	}

	processingReporter.finish("OSV PostgreSQL stage completed")

	osvLog.Printf("OSV replication completed: ecosystem=%s records=%d", replicator.payload.Ecosystem, saved.Load())

	if err := replicator.ctx.Err(); err != nil {
		return nil, err
	}

	return newestModifiedAt, nil
}

func (replicator *Replicator) runIncremental() error {
	ctx := replicator.ctx
	payload := replicator.payload
	state := replicator.state
	queue, err := redisGetOsvQueue(ctx, payload.Ecosystem)
	if err != nil {
		return fmt.Errorf("get OSV incremental queue: %w", err)
	}
	if len(queue) > 0 {
		osvLog.Printf("OSV incremental queue pending: ecosystem=%s, pending=%d, processing one vulnerability", payload.Ecosystem, len(queue))
		return replicator.processIncrementalQueueItem(queue)
	}

	nextCheckAt := time.Now().Add(osvReplicationCooldown)
	claimed, err := dbClaimOsvReplicationCheck(ctx, payload.Ecosystem, nextCheckAt)
	if err != nil {
		return fmt.Errorf("claim OSV incremental replication: %w", err)
	}
	if !claimed {
		osvLog.Printf("OSV incremental synchronization skipped: ecosystem=%s, cooldown active, next check at %s, last check at %s", payload.Ecosystem, formatOsvTime(state.NextCheckAt), formatOsvTime(state.LastCheckedAt))
		return nil
	}
	osvLog.Printf("OSV cooldown started: ecosystem=%s, mode=incremental, reason=modified feed download, next check at %s", payload.Ecosystem, formatOsvTime(&nextCheckAt))
	osvLog.Printf("OSV incremental synchronization: ecosystem=%s, downloading modified feed", payload.Ecosystem)

	feed, err := fetchOSVModifiedIDCSV(ctx, payload.Ecosystem)
	if err != nil {
		osvLog.Printf("OSV cooldown retained: ecosystem=%s, mode=incremental, reason=modified feed request failed, next check at %s, error=%v", payload.Ecosystem, formatOsvTime(&nextCheckAt), err)
		return fmt.Errorf("fetch OSV modified feed: %w", err)
	}
	defer feed.Close()

	changes, newestModifiedAt, err := parseModifiedIDCSV(feed, state.LastModifiedAt)
	if err != nil {
		osvLog.Printf("OSV cooldown retained: ecosystem=%s, mode=incremental, reason=modified feed parsing failed, next check at %s, error=%v", payload.Ecosystem, formatOsvTime(&nextCheckAt), err)
		return fmt.Errorf("parse OSV modified feed: %w", err)
	}
	osvLog.Printf("OSV modified feed synchronized: ecosystem=%s, changes=%d, newest_modified_at=%v", payload.Ecosystem, len(changes), newestModifiedAt)
	if len(changes) == 0 {
		osvLog.Printf("OSV incremental synchronization found no new changes: ecosystem=%s, cooldown until %s", payload.Ecosystem, formatOsvTime(&nextCheckAt))
	}

	queue = mergeOsvQueue(queue, changes)
	if err := redisSetOsvQueue(ctx, payload.Ecosystem, queue); err != nil {
		osvLog.Printf("OSV cooldown retained: ecosystem=%s, mode=incremental, reason=queue persistence failed, next check at %s, error=%v", payload.Ecosystem, formatOsvTime(&nextCheckAt), err)
		return fmt.Errorf("save OSV incremental queue: %w", err)
	}

	if err := dbCompleteOsvIncrementalReplication(ctx, payload.Ecosystem, newestModifiedAt); err != nil {
		osvLog.Printf("OSV cooldown retained: ecosystem=%s, mode=incremental, reason=state update failed, next check at %s, error=%v", payload.Ecosystem, formatOsvTime(&nextCheckAt), err)
		return fmt.Errorf("complete OSV incremental replication: %w", err)
	}
	osvLog.Printf("OSV incremental synchronization completed: ecosystem=%s, changes=%d, pending=%d, cooldown until %s", payload.Ecosystem, len(changes), len(queue), formatOsvTime(&nextCheckAt))
	return nil
}

func (replicator *Replicator) processIncrementalQueueItem(queue []ModifiedVulnerability) error {
	ctx := replicator.ctx
	payload := replicator.payload
	current := queue[0]
	osvLog.Printf("OSV incremental synchronization: ecosystem=%s, synchronizing vulnerability=%s, remaining=%d", payload.Ecosystem, current.ID, len(queue)-1)

	vulnerability, status, err := fetchOSVVulnerability(ctx, payload.Ecosystem, current.ID)
	if err != nil {
		return err
	}

	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("start OSV incremental transaction: %w", err)
	}
	defer transaction.Rollback(ctx)

	if status == 404 {
		if err := dbMarkVulnerabilityWithdrawn(ctx, transaction, current.ID, current.ModifiedAt); err != nil {
			return fmt.Errorf("mark OSV vulnerability withdrawn: %w", err)
		}
	} else {
		parsedVulnerability := parseVulnerability(*vulnerability)
		if err := dbSaveVulnerabilityTx(ctx, transaction, parsedVulnerability); err != nil {
			return fmt.Errorf("save OSV vulnerability: %w", err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit OSV incremental transaction: %w", err)
	}
	if err := redisSetOsvQueue(ctx, payload.Ecosystem, queue[1:]); err != nil {
		return fmt.Errorf("advance OSV incremental queue: %w", err)
	}

	resultStatus := "updated"
	if status == 404 {
		resultStatus = "withdrawn"
	}
	osvLog.Printf("OSV incremental vulnerability completed: ecosystem=%s, vulnerability=%s, status=%s, remaining=%d", payload.Ecosystem, current.ID, resultStatus, len(queue)-1)
	if len(queue) == 1 {
		osvLog.Printf("OSV incremental queue drained: ecosystem=%s, next callback will check the modified feed after its cooldown", payload.Ecosystem)
	}
	return nil
}

func mergeOsvQueue(queue []ModifiedVulnerability, changes []ModifiedVulnerability) []ModifiedVulnerability {
	indexes := make(map[string]int, len(queue)+len(changes))
	merged := append([]ModifiedVulnerability(nil), queue...)
	for index, item := range merged {
		indexes[item.ID] = index
	}
	for _, change := range changes {
		if index, exists := indexes[change.ID]; exists {
			if change.ModifiedAt.After(merged[index].ModifiedAt) {
				merged[index] = change
			}
			continue
		}
		indexes[change.ID] = len(merged)
		merged = append(merged, change)
	}
	return merged
}

func formatOsvTime(value *time.Time) string {
	if value == nil {
		return "not scheduled"
	}
	return value.Format("2006-01-02 15:04:05 -0700 MST")
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func waitForOSVWritePause(ctx context.Context) bool {
	timer := time.NewTimer(osvWritePause)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

type byteReaderAt []byte

func (data byteReaderAt) ReadAt(target []byte, offset int64) (int, error) {
	if offset >= int64(len(data)) {
		return 0, io.EOF
	}
	read := copy(target, data[offset:])
	if read < len(target) {
		return read, io.EOF
	}
	return read, nil
}
