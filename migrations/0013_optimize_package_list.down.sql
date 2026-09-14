BEGIN;

DROP INDEX IF EXISTS idx_vulnerability_affected_package_vulnerability;

CREATE INDEX idx_vulnerability_affected_package_id
    ON vulnerability_affected (package_id);

DROP INDEX IF EXISTS idx_packages_data_popularity_package_id;
DROP INDEX IF EXISTS idx_packages_ecosystem_lower_name_id;

COMMIT;