BEGIN;

ALTER TABLE vulnerabilities
    ADD COLUMN withdrawn_at TIMESTAMPTZ;

CREATE TABLE osv_replication_state (
    ecosystem TEXT PRIMARY KEY,
    is_first_synchronised BOOLEAN NOT NULL DEFAULT FALSE,
    last_modified_at TIMESTAMPTZ,
    last_checked_at TIMESTAMPTZ,
    next_check_at TIMESTAMPTZ,
    last_synchronised_at TIMESTAMPTZ
);

COMMIT;