INSERT INTO packages_data (
	package_id,
	is_deprecated,
	need_update,
	popularity,
	updated_at
)
SELECT DISTINCT
	packages.id,
	TRUE,
	TRUE,
	0,
	NOW()
FROM packages
WHERE packages.ecosystem = 'npm'
  AND packages.name = ANY($1::text[])
ON CONFLICT (package_id) DO UPDATE SET
	is_deprecated = TRUE,
	need_update = TRUE,
	popularity = 0,
	updated_at = NOW();
