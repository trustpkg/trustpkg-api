CREATE TABLE IF NOT EXISTS downloads (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    version_id BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    download_count BIGINT NOT NULL DEFAULT 0,
    period_from TIMESTAMP NOT NULL,
    period_to TIMESTAMP NOT NULL,

    UNIQUE (version_id, period_from, period_to)
);