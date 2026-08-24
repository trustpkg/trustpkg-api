BEGIN;

ALTER TABLE packages
    ADD COLUMN ecosystem TEXT NOT NULL DEFAULT 'npm';

ALTER TABLE packages
    DROP CONSTRAINT packages_name_key;

ALTER TABLE packages
    ADD CONSTRAINT packages_ecosystem_name_key
    UNIQUE (ecosystem, name);

ALTER TABLE packages
    ALTER COLUMN ecosystem DROP DEFAULT;


ALTER TABLE vulnerability_affected
    ADD COLUMN package_id BIGINT
        REFERENCES packages(id)
        ON DELETE CASCADE;

CREATE INDEX idx_vulnerability_affected_package_id
    ON vulnerability_affected(package_id);

ALTER TABLE vulnerability_affected
    DROP CONSTRAINT IF EXISTS
    vulnerability_affected_vulnerability_id_ecosystem_package_n_key;

ALTER TABLE vulnerability_affected
    DROP COLUMN ecosystem,
    DROP COLUMN package_name;

ALTER TABLE vulnerability_affected
    ADD CONSTRAINT vulnerability_affected_unique
    UNIQUE (
        vulnerability_id,
        package_id,
        introduced_version,
        fixed_version,
        last_affected_version
    );

COMMIT;