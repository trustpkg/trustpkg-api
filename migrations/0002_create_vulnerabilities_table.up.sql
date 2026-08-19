CREATE TABLE IF NOT EXISTS vulnerabilities (
    id BIGSERIAL PRIMARY KEY,
    osv_id TEXT NOT NULL UNIQUE,
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
    reference_urls TEXT[]
);

CREATE TABLE IF NOT EXISTS package_vulnerabilities (
    package_id BIGINT NOT NULL
        REFERENCES packages(id) ON DELETE CASCADE,

    vulnerability_id BIGINT NOT NULL
        REFERENCES vulnerabilities(id) ON DELETE CASCADE,

    PRIMARY KEY (package_id, vulnerability_id)
);