BEGIN;

CREATE TABLE typesense_state (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    phase TEXT NOT NULL,
    watchlist_offset INTEGER NOT NULL DEFAULT 0,
    last_package_id BIGINT NOT NULL DEFAULT 0,
    cycle BIGINT NOT NULL DEFAULT 0,
    documents_uploaded BIGINT NOT NULL DEFAULT 0,
    watchlist_documents_uploaded BIGINT NOT NULL DEFAULT 0,
    database_documents_uploaded BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'idle',
    last_error TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
