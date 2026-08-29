UPDATE osv_replication_state
SET is_first_synchronised = TRUE,
    last_modified_at = COALESCE($2, last_modified_at),
    next_check_at = CURRENT_TIMESTAMP + INTERVAL '10 minutes',
    last_synchronised_at = CURRENT_TIMESTAMP
WHERE ecosystem = $1;