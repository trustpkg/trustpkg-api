BEGIN;

DROP TABLE IF EXISTS osv_replication_state;

ALTER TABLE vulnerabilities
    DROP COLUMN IF EXISTS withdrawn_at;

COMMIT;