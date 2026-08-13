CREATE TABLE IF NOT EXISTS versions (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    description TEXT,
    license TEXT,
    published_at TIMESTAMP DEFAULT NULL,
    dependencies TEXT[] DEFAULT NULL,
    optional_dependencies TEXT[] DEFAULT NULL,
    dev_dependencies TEXT[] DEFAULT NULL,
    peer_dependencies TEXT[] DEFAULT NULL
);