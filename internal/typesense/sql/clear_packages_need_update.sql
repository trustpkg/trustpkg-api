UPDATE packages_data
SET need_update = FALSE
WHERE package_id = ANY($1::bigint[]);