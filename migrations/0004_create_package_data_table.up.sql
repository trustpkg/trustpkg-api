CREATE TABLE IF NOT EXISTS packages_data (
    package_id BIGINT PRIMARY KEY
        REFERENCES packages(id)
        ON DELETE CASCADE,
    is_deprecated BOOLEAN NOT NULL DEFAULT FALSE,
    need_update BOOLEAN NOT NULL DEFAULT FALSE,
    popularity INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);