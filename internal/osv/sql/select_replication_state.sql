SELECT ecosystem,
       is_first_synchronised,
       last_modified_at,
       last_checked_at,
       next_check_at,
       last_synchronised_at
FROM osv_replication_state
WHERE ecosystem = $1;