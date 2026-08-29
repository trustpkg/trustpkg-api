WITH claimed AS (
    UPDATE osv_replication_state
    SET last_checked_at = CURRENT_TIMESTAMP,
        next_check_at = $2
    WHERE ecosystem = $1
      AND (next_check_at IS NULL OR next_check_at <= CURRENT_TIMESTAMP)
    RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM claimed);