BEGIN;

DROP INDEX IF EXISTS idx_packages_slug;
DROP INDEX IF EXISTS idx_packages_slug_lookup;

DROP TRIGGER IF EXISTS packages_set_slug_before_write ON packages;

ALTER TABLE packages
    DROP COLUMN IF EXISTS slug;

DROP FUNCTION IF EXISTS set_package_slug();
DROP FUNCTION IF EXISTS normalize_package_slug(TEXT);

COMMIT;
