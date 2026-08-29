INSERT INTO vulnerability_affected (
    vulnerability_id, package_id, introduced_version,
    fixed_version, last_affected_version
)
SELECT $1, $2, $3, $4, $5
WHERE NOT EXISTS (
    SELECT 1
    FROM vulnerability_affected
    WHERE vulnerability_id = $1
      AND package_id = $2
      AND introduced_version IS NOT DISTINCT FROM $3
      AND fixed_version IS NOT DISTINCT FROM $4
      AND last_affected_version IS NOT DISTINCT FROM $5
);
