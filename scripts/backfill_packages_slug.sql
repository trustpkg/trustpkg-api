\set ON_ERROR_STOP on

\if :{?batch_size}
\else
\set batch_size 2000
\endif

\if :{?pause_seconds}
\else
\set pause_seconds 0.20
\endif

SELECT EXISTS (
    SELECT 1
    FROM pg_class index_class
    JOIN pg_index index_metadata ON index_metadata.indexrelid = index_class.oid
    JOIN information_schema.columns columns
        ON columns.table_name = 'packages'
       AND columns.column_name = 'slug'
       AND columns.is_nullable = 'NO'
    WHERE index_class.relname = 'idx_packages_slug'
      AND index_class.relkind = 'i'
      AND index_metadata.indisunique
      AND index_metadata.indisvalid
) AS slug_backfill_complete \gset

\if :slug_backfill_complete
\echo 'Package slug backfill is already complete.'
DROP INDEX CONCURRENTLY IF EXISTS idx_packages_slug_lookup;
DROP TABLE IF EXISTS package_slug_backfill_candidates;
DROP TABLE IF EXISTS package_slug_backfill_state;
DROP PROCEDURE IF EXISTS backfill_package_slugs(INTEGER, DOUBLE PRECISION);
DROP PROCEDURE IF EXISTS backfill_package_slugs(INTEGER);
\quit
\endif

CREATE TABLE IF NOT EXISTS package_slug_backfill_candidates (
    id BIGINT PRIMARY KEY,
    normalized_slug TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS package_slug_backfill_candidates_slug_idx
    ON package_slug_backfill_candidates (normalized_slug, id);

CREATE TABLE IF NOT EXISTS package_slug_backfill_state (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    candidates_last_id BIGINT NOT NULL DEFAULT 0,
    slugs_last_id BIGINT NOT NULL DEFAULT 0,
    deduplicate_last_id BIGINT NOT NULL DEFAULT 0,
    candidates_processed BIGINT NOT NULL DEFAULT 0,
    slugs_processed BIGINT NOT NULL DEFAULT 0,
    deduplicated_processed BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

ALTER TABLE package_slug_backfill_state
    ADD COLUMN IF NOT EXISTS deduplicate_last_id BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS candidates_processed BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS slugs_processed BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS deduplicated_processed BIGINT NOT NULL DEFAULT 0;

INSERT INTO package_slug_backfill_state (id)
VALUES (TRUE)
ON CONFLICT (id) DO NOTHING;

CREATE OR REPLACE PROCEDURE backfill_package_slugs(
    batch_size INTEGER DEFAULT 2000,
    pause_seconds DOUBLE PRECISION DEFAULT 0.20
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_candidates_last_id BIGINT;
    v_slugs_last_id BIGINT;
    v_deduplicate_last_id BIGINT;
    v_batch_last_id BIGINT;
    v_max_package_id BIGINT;
    v_batch_count BIGINT;
    v_candidates_processed BIGINT;
    v_slugs_processed BIGINT;
    v_deduplicated_processed BIGINT;
    package_record RECORD;
    v_candidate_slug TEXT;
    v_batch_started_at TIMESTAMPTZ;
BEGIN
    IF batch_size < 1 THEN
        RAISE EXCEPTION 'batch_size must be greater than zero';
    END IF;

    SELECT COALESCE(MAX(id), 0)
    INTO v_max_package_id
    FROM packages;

    RAISE NOTICE '[start] batch_size=%, pause_seconds=%, max_package_id=%, at=%',
        batch_size, pause_seconds, v_max_package_id, clock_timestamp();

    LOOP
        v_batch_started_at := clock_timestamp();

        SELECT state.candidates_last_id, state.candidates_processed
        INTO v_candidates_last_id, v_candidates_processed
        FROM package_slug_backfill_state state
        WHERE state.id = TRUE;

        SELECT MAX(batch.id), COUNT(*)
        INTO v_batch_last_id, v_batch_count
        FROM (
            SELECT packages.id
            FROM packages
            WHERE packages.id > v_candidates_last_id
            ORDER BY packages.id
            LIMIT batch_size
        ) batch;

        EXIT WHEN v_batch_last_id IS NULL;

        INSERT INTO package_slug_backfill_candidates (id, normalized_slug)
        SELECT packages.id, normalize_package_slug(packages.name)
        FROM packages
                WHERE packages.id > v_candidates_last_id
                    AND packages.id <= v_batch_last_id
        ON CONFLICT (id) DO NOTHING;

        v_candidates_processed := v_candidates_processed + v_batch_count;

        UPDATE package_slug_backfill_state
        SET candidates_last_id = v_batch_last_id,
            candidates_processed = v_candidates_processed,
            updated_at = clock_timestamp()
        WHERE id = TRUE;

        COMMIT;

        RAISE NOTICE '[candidates] last_id=%, batch=%, processed=%, progress=%.2f%%, duration=%, at=%',
            v_batch_last_id,
            v_batch_count,
            v_candidates_processed,
            CASE
                WHEN v_max_package_id = 0 THEN 100
                ELSE round((100.0 * v_batch_last_id / v_max_package_id)::numeric, 2)
            END,
            clock_timestamp() - v_batch_started_at,
            clock_timestamp();

        PERFORM pg_sleep(GREATEST(pause_seconds, 0));
    END LOOP;

    RAISE NOTICE '[candidates] complete at=%', clock_timestamp();

    LOOP
        v_batch_started_at := clock_timestamp();

        SELECT state.slugs_last_id, state.slugs_processed
        INTO v_slugs_last_id, v_slugs_processed
        FROM package_slug_backfill_state state
        WHERE state.id = TRUE;

        SELECT MAX(batch.id), COUNT(*)
        INTO v_batch_last_id, v_batch_count
        FROM (
            SELECT candidates.id
            FROM package_slug_backfill_candidates candidates
            WHERE candidates.id > v_slugs_last_id
            ORDER BY candidates.id
            LIMIT batch_size
        ) batch;

        EXIT WHEN v_batch_last_id IS NULL;

        UPDATE packages
        SET slug = CASE
            WHEN first_package.id = candidates.id
                THEN candidates.normalized_slug
            ELSE candidates.normalized_slug || '-' || packages.id
        END
        FROM package_slug_backfill_candidates candidates
        JOIN LATERAL (
            SELECT MIN(first_candidate.id) AS id
            FROM package_slug_backfill_candidates first_candidate
            WHERE first_candidate.normalized_slug = candidates.normalized_slug
        ) first_package ON TRUE
        WHERE packages.id = candidates.id
          AND candidates.id > v_slugs_last_id
          AND candidates.id <= v_batch_last_id;

        v_slugs_processed := v_slugs_processed + v_batch_count;

        UPDATE package_slug_backfill_state
        SET slugs_last_id = v_batch_last_id,
            slugs_processed = v_slugs_processed,
            updated_at = clock_timestamp()
        WHERE id = TRUE;

        COMMIT;

        RAISE NOTICE '[slugs] last_id=%, batch=%, processed=%, progress=%.2f%%, duration=%, at=%',
            v_batch_last_id,
            v_batch_count,
            v_slugs_processed,
            CASE
                WHEN v_max_package_id = 0 THEN 100
                ELSE round((100.0 * v_batch_last_id / v_max_package_id)::numeric, 2)
            END,
            clock_timestamp() - v_batch_started_at,
            clock_timestamp();

        PERFORM pg_sleep(GREATEST(pause_seconds, 0));
    END LOOP;

    RAISE NOTICE '[slugs] complete at=%', clock_timestamp();

    LOOP
        v_batch_started_at := clock_timestamp();
        v_deduplicated_processed := 0;

        SELECT state.deduplicate_last_id, state.deduplicated_processed
        INTO v_deduplicate_last_id, v_deduplicated_processed
        FROM package_slug_backfill_state state
        WHERE state.id = TRUE;

        v_batch_last_id := v_deduplicate_last_id;
        v_batch_count := 0;

        FOR package_record IN
            SELECT packages.id, packages.slug
            FROM packages
            WHERE packages.id > v_deduplicate_last_id
              AND packages.slug IS NOT NULL
              AND EXISTS (
                  SELECT 1
                  FROM packages previous_package
                  WHERE previous_package.slug = packages.slug
                    AND previous_package.id < packages.id
              )
            ORDER BY packages.id
            LIMIT batch_size
        LOOP
            v_batch_last_id := package_record.id;
            v_batch_count := v_batch_count + 1;
            v_candidate_slug := package_record.slug || '-' || package_record.id;

            WHILE EXISTS (
                SELECT 1
                FROM packages
                WHERE packages.slug = v_candidate_slug
                  AND packages.id <> package_record.id
            ) LOOP
                v_candidate_slug := v_candidate_slug || '-1';
            END LOOP;

            UPDATE packages
            SET slug = v_candidate_slug
            WHERE packages.id = package_record.id;
        END LOOP;

        EXIT WHEN v_batch_count = 0;

        v_deduplicated_processed := v_deduplicated_processed + v_batch_count;

        UPDATE package_slug_backfill_state
        SET deduplicate_last_id = v_batch_last_id,
            deduplicated_processed = v_deduplicated_processed,
            updated_at = clock_timestamp()
        WHERE id = TRUE;

        COMMIT;

        RAISE NOTICE '[deduplicate] last_id=%, batch=%, processed=%, progress=%.2f%%, duration=%, at=%',
            v_batch_last_id,
            v_batch_count,
            v_deduplicated_processed,
            CASE
                WHEN v_max_package_id = 0 THEN 100
                ELSE round((100.0 * v_batch_last_id / v_max_package_id)::numeric, 2)
            END,
            clock_timestamp() - v_batch_started_at,
            clock_timestamp();

        PERFORM pg_sleep(GREATEST(pause_seconds, 0));
    END LOOP;

    RAISE NOTICE '[deduplicate] complete at=%', clock_timestamp();
END;
$$;

\echo 'Starting resumable package slug backfill. Stop with Ctrl-C; rerun the same command to resume.'
CALL backfill_package_slugs(:batch_size, :pause_seconds);

ALTER TABLE packages
    ALTER COLUMN slug SET NOT NULL;

DROP INDEX CONCURRENTLY IF EXISTS idx_packages_slug;

CREATE UNIQUE INDEX CONCURRENTLY idx_packages_slug
    ON packages (slug);

DROP INDEX CONCURRENTLY IF EXISTS idx_packages_slug_lookup;
DROP TABLE package_slug_backfill_candidates;
DROP TABLE package_slug_backfill_state;
DROP PROCEDURE backfill_package_slugs(INTEGER, DOUBLE PRECISION);
DROP PROCEDURE IF EXISTS backfill_package_slugs(INTEGER);

SELECT
    COUNT(*) AS packages,
    COUNT(slug) AS packages_with_slug,
    COUNT(DISTINCT slug) AS distinct_slugs
FROM packages;
