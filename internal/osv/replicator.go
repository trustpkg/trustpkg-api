package osv

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"
)

const maxOSVWorkers = 12

func ReplicateEcosystem(ctx context.Context, payload FetchOSVEcosystemDumpPayload) error {
	dump, err := FetchOSVEcosystemDump(payload)
	if err != nil {
		return fmt.Errorf("fetch OSV dump: %w", err)
	}
	archiveData, err := io.ReadAll(dump)
	dump.Close()
	if err != nil {
		return fmt.Errorf("read OSV dump: %w", err)
	}
	archive, err := zip.NewReader(byteReaderAt(archiveData), int64(len(archiveData)))
	if err != nil {
		return fmt.Errorf("open OSV ZIP: %w", err)
	}

	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	processingReporter := newStatusReporter("OSV PostgreSQL", "records", int64(len(archive.File)), 1)
	jobs := make(chan ParsedVulnerability, maxOSVWorkers)
	var workers sync.WaitGroup
	var saved atomic.Int64
	var firstErr error
	var errorOnce sync.Once
	for worker := 0; worker < maxOSVWorkers; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for vulnerability := range jobs {
				if err := saveVulnerability(workerContext, vulnerability); err != nil {
					errorOnce.Do(func() {
						firstErr = err
						cancel()
					})
				} else {
					processingReporter.update(saved.Add(1))
				}
			}
		}()
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
			errorOnce.Do(func() { firstErr = fmt.Errorf("decode %s: %w", file.Name, decodeErr) })
			break
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
		return firstErr
	}
	processingReporter.finish("OSV PostgreSQL stage completed")
	log.Printf("OSV replication completed: ecosystem=%s records=%d", payload.Ecosystem, saved.Load())
	return ctx.Err()
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
