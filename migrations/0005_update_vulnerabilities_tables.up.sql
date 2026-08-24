BEGIN;

DROP TABLE IF EXISTS package_vulnerabilities;
DROP TABLE IF EXISTS vulnerabilities;

CREATE TABLE vulnerabilities (
    id BIGSERIAL PRIMARY KEY,
    osv_id TEXT NOT NULL UNIQUE,
    cve_id TEXT,
    summary TEXT,
    description TEXT,
    severity TEXT,
    cvss_score NUMERIC(3, 1),
    cvss_vector TEXT,
    published_at TIMESTAMP,
    modified_at TIMESTAMP
);

CREATE TABLE vulnerability_affected (
    id BIGSERIAL PRIMARY KEY,

    vulnerability_id BIGINT NOT NULL
        REFERENCES vulnerabilities(id)
        ON DELETE CASCADE,

    ecosystem TEXT NOT NULL,
    package_name TEXT NOT NULL,

    introduced_version TEXT,
    fixed_version TEXT,
    last_affected_version TEXT,

    UNIQUE (
        vulnerability_id,
        ecosystem,
        package_name,
        introduced_version,
        fixed_version,
        last_affected_version
    )
);

CREATE INDEX idx_vulnerability_affected_package
    ON vulnerability_affected (ecosystem, package_name);

CREATE INDEX idx_vulnerability_affected_vulnerability
    ON vulnerability_affected (vulnerability_id);

CREATE TABLE vulnerability_references (
    id BIGSERIAL PRIMARY KEY,

    vulnerability_id BIGINT NOT NULL
        REFERENCES vulnerabilities(id)
        ON DELETE CASCADE,

    url TEXT NOT NULL,

    UNIQUE (vulnerability_id, url)
);

CREATE INDEX idx_vulnerability_references_vulnerability
    ON vulnerability_references (vulnerability_id);

COMMIT;