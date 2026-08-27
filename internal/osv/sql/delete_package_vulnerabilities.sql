DELETE FROM vulnerability_affected
WHERE package_id = (
    SELECT id
    FROM packages
    WHERE name = $1 AND ecosystem = $2
);