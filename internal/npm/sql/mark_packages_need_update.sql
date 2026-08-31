INSERT INTO packages_data (package_id, need_update)
SELECT packages.id, FALSE
FROM packages
WHERE packages.ecosystem = 'npm'
  AND packages.name = ANY($1::text[])
ON CONFLICT (package_id) DO UPDATE
SET need_update = FALSE;
