package typesense

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultSlugReindexCheckpointPath = "tmp/typesense-packages-slug-reindex.json"

type SlugReindexOptions struct {
	BatchSize      int
	CheckpointPath string
	Pause          time.Duration
	Reset          bool
}

type slugReindexCheckpoint struct {
	LastPackageID     int64 `json:"last_package_id"`
	DocumentsUploaded int64 `json:"documents_uploaded"`
	Completed         bool  `json:"completed"`
}

func ReindexPackageSlugs(ctx context.Context, writer TypesenseWriter, options SlugReindexOptions) error {
	if writer == nil {
		return fmt.Errorf("Typesense writer cannot be nil")
	}
	if options.BatchSize <= 0 {
		options.BatchSize = pipelineBatchSize
	}
	if options.CheckpointPath == "" {
		options.CheckpointPath = defaultSlugReindexCheckpointPath
	}

	checkpoint, err := loadSlugReindexCheckpoint(options.CheckpointPath, options.Reset)
	if err != nil {
		return err
	}
	if checkpoint.Completed {
		typesenseLog.Printf("Typesense slug reindex already completed: uploaded=%d", checkpoint.DocumentsUploaded)
		return nil
	}

	typesenseLog.Printf("Typesense slug reindex started: last_package_id=%d batch_size=%d", checkpoint.LastPackageID, options.BatchSize)
	for {
		documents, err := dbSelectSlugReindexDocuments(ctx, checkpoint.LastPackageID, options.BatchSize)
		if err != nil {
			return fmt.Errorf("select slug reindex documents after package %d: %w", checkpoint.LastPackageID, err)
		}
		if len(documents) == 0 {
			checkpoint.Completed = true
			if err := saveSlugReindexCheckpoint(options.CheckpointPath, checkpoint); err != nil {
				return err
			}
			typesenseLog.Printf("Typesense slug reindex completed: uploaded=%d last_package_id=%d", checkpoint.DocumentsUploaded, checkpoint.LastPackageID)
			return nil
		}

		payload := make([]any, len(documents))
		for index := range documents {
			payload[index] = documents[index]
		}
		if err := writer.UpsertDocuments(&UpsertPayload{
			Name:      PackageSchema.Name,
			BatchSize: options.BatchSize,
			Documents: payload,
		}); err != nil {
			return fmt.Errorf("upload slug reindex batch after package %d: %w", checkpoint.LastPackageID, err)
		}

		checkpoint.LastPackageID = packageID(documents[len(documents)-1].ID)
		checkpoint.DocumentsUploaded += int64(len(documents))
		if err := saveSlugReindexCheckpoint(options.CheckpointPath, checkpoint); err != nil {
			return err
		}
		typesenseLog.Printf("Typesense slug reindex progress: uploaded=%d last_package_id=%d", checkpoint.DocumentsUploaded, checkpoint.LastPackageID)

		if options.Pause > 0 {
			timer := time.NewTimer(options.Pause)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
}

func loadSlugReindexCheckpoint(path string, reset bool) (slugReindexCheckpoint, error) {
	if reset {
		return slugReindexCheckpoint{}, nil
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return slugReindexCheckpoint{}, nil
	}
	if err != nil {
		return slugReindexCheckpoint{}, fmt.Errorf("read slug reindex checkpoint %q: %w", path, err)
	}

	var checkpoint slugReindexCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return slugReindexCheckpoint{}, fmt.Errorf("decode slug reindex checkpoint %q: %w", path, err)
	}
	return checkpoint, nil
}

func saveSlugReindexCheckpoint(path string, checkpoint slugReindexCheckpoint) error {
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("encode slug reindex checkpoint: %w", err)
	}

	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create slug reindex checkpoint directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".typesense-packages-slug-reindex-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary slug reindex checkpoint: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write slug reindex checkpoint: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close slug reindex checkpoint: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace slug reindex checkpoint: %w", err)
	}
	return nil
}
