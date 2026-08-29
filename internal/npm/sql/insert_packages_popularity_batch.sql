WITH package_popularity AS (
	SELECT
		key AS package_name,
		value::BIGINT AS popularity
	FROM jsonb_each_text($1::jsonb)
)
INSERT INTO packages_data (
	package_id,
	is_deprecated,
	need_update,
	popularity,
	updated_at
)
SELECT
	packages.id,
	FALSE,
	TRUE,
	package_popularity.popularity,
	NOW()
FROM packages
JOIN package_popularity
	ON package_popularity.package_name = packages.name
WHERE packages.ecosystem = 'npm'
ON CONFLICT (package_id) DO UPDATE SET
	need_update = TRUE,
	popularity = EXCLUDED.popularity,
	updated_at = NOW();
