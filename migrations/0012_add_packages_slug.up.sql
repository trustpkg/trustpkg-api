BEGIN;

ALTER TABLE packages
    ADD COLUMN slug TEXT;

CREATE OR REPLACE FUNCTION normalize_package_slug(package_name TEXT)
RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
STRICT
AS $$
    SELECT regexp_replace(
        regexp_replace(
            regexp_replace(
                regexp_replace(
                    regexp_replace(package_name, '%40', '', 'gi'),
                    '%2f', '-', 'gi'
                ),
                '%2e', '-', 'gi'
            ),
            '@', '', 'g'
        ),
        '[/.]', '-', 'g'
    );
$$;

CREATE OR REPLACE FUNCTION set_package_slug()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    normalized_slug TEXT;
    candidate_slug TEXT;
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.name IS NOT DISTINCT FROM OLD.name THEN
        RETURN NEW;
    END IF;

    normalized_slug := normalize_package_slug(NEW.name);
    candidate_slug := normalized_slug;

    PERFORM pg_advisory_xact_lock(hashtextextended(normalized_slug, 0));

    IF EXISTS (
        SELECT 1
        FROM packages
        WHERE slug = candidate_slug
          AND id <> NEW.id
    ) THEN
        candidate_slug := normalized_slug || '-' || NEW.id;

        WHILE EXISTS (
            SELECT 1
            FROM packages
            WHERE slug = candidate_slug
              AND id <> NEW.id
        ) LOOP
            candidate_slug := candidate_slug || '-1';
        END LOOP;
    END IF;

    NEW.slug := candidate_slug;
    RETURN NEW;
END;
$$;

CREATE TRIGGER packages_set_slug_before_write
BEFORE INSERT OR UPDATE OF name ON packages
FOR EACH ROW
EXECUTE FUNCTION set_package_slug();

CREATE INDEX idx_packages_slug_lookup
    ON packages (slug)
    WHERE slug IS NOT NULL;

COMMIT;

