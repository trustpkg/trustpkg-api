BEGIN;

CREATE INDEX idx_packages_ecosystem_lower_name_id
    ON packages (ecosystem, LOWER(name), id);

CREATE INDEX idx_packages_data_popularity_package_id
    ON packages_data (popularity DESC, package_id);

DROP INDEX IF EXISTS idx_vulnerability_affected_package_id;

CREATE INDEX idx_vulnerability_affected_package_vulnerability
    ON vulnerability_affected (package_id, vulnerability_id);

COMMIT;