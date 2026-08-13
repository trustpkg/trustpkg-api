CREATE TABLE IF NOT EXISTS vulnerabilities (
    id BIGSERIAL PRIMARY KEY,
    version_id BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    osv_id TEXT NOT NULL,
    cve_id TEXT,
    summary TEXT,
    description TEXT,
    severity TEXT,
    cvss_score NUMERIC(3, 1),
    cvss_vector TEXT,
    affected_ranges TEXT[],
    fixed_versions TEXT[],
    published_at TIMESTAMP,
    modified_at TIMESTAMP,
    reference_urls TEXT[],
    UNIQUE (osv_id)
);