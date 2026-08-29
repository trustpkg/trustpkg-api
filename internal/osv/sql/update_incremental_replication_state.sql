UPDATE osv_replication_state
SET last_modified_at = COALESCE($2, last_modified_at),
    last_synchronised_at = CURRENT_TIMESTAMP
WHERE ecosystem = $1;