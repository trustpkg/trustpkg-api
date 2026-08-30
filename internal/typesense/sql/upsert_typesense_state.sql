INSERT INTO typesense_state (
    id, phase, watchlist_offset, last_package_id, cycle,
    documents_uploaded, watchlist_documents_uploaded,
    database_documents_uploaded, status, last_error,
    started_at, finished_at, updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''),
        CASE WHEN $8 = 'running' THEN COALESCE((SELECT started_at FROM typesense_state WHERE id = 1), CURRENT_TIMESTAMP) ELSE (SELECT started_at FROM typesense_state WHERE id = 1) END,
        CASE WHEN $8 IN ('completed', 'failed') THEN CURRENT_TIMESTAMP ELSE NULL END,
        CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET
    phase = EXCLUDED.phase,
    watchlist_offset = EXCLUDED.watchlist_offset,
    last_package_id = EXCLUDED.last_package_id,
    cycle = EXCLUDED.cycle,
    documents_uploaded = EXCLUDED.documents_uploaded,
    watchlist_documents_uploaded = EXCLUDED.watchlist_documents_uploaded,
    database_documents_uploaded = EXCLUDED.database_documents_uploaded,
    status = EXCLUDED.status,
    last_error = EXCLUDED.last_error,
    started_at = COALESCE(typesense_state.started_at, EXCLUDED.started_at),
    finished_at = EXCLUDED.finished_at,
    updated_at = CURRENT_TIMESTAMP;
