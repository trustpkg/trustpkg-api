BEGIN;

ALTER TABLE vulnerability_affected
    DROP CONSTRAINT vulnerability_affected_unique;

ALTER TABLE vulnerability_affected
    ADD COLUMN ecosystem TEXT NOT NULL DEFAULT 'npm';

ALTER TABLE vulnerability_affected
    ADD COLUMN package_name TEXT NOT NULL DEFAULT '';

ALTER TABLE vulnerability_affected
    DROP CONSTRAINT vulnerability_affected_package_id_fkey;

DROP INDEX IF EXISTS idx_vulnerability_affected_package_id;

ALTER TABLE vulnerability_affected
    DROP COLUMN package_id;

ALTER TABLE vulnerability_affected
    ADD CONSTRAINT vulnerability_affected_vulnerability_id_ecosystem_package_name_introduced_version_fixed_version_last_affected_version_key
    UNIQUE (
        vulnerability_id,
        ecosystem,
        package_name,
        introduced_version,
        fixed_version,
        last_affected_version
    );

ALTER TABLE vulnerability_affected
    ALTER COLUMN ecosystem DROP DEFAULT;

ALTER TABLE vulnerability_affected
    ALTER COLUMN package_name DROP DEFAULT;

ALTER TABLE packages
    DROP CONSTRAINT packages_ecosystem_name_key;

ALTER TABLE packages
    ADD CONSTRAINT packages_name_key
    UNIQUE (name);

ALTER TABLE packages
    DROP COLUMN ecosystem;

COMMIT;