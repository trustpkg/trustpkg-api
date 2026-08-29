WITH packages_without_data AS (
	SELECT
		packages.id,
		packages.name
	FROM packages
	LEFT JOIN packages_data
		ON packages_data.package_id = packages.id
	WHERE packages_data.package_id IS NULL
		AND packages.name NOT LIKE '@%'
	ORDER BY packages.id
	LIMIT 128
), packages_with_old_data AS (
	SELECT
		packages.id,
		packages.name
	FROM packages
	JOIN packages_data
		ON packages_data.package_id = packages.id
	WHERE packages.name NOT LIKE '@%'
		AND packages_data.updated_at <= CURRENT_TIMESTAMP - INTERVAL '1 month'
	ORDER BY packages.id
	LIMIT GREATEST(128 - (SELECT COUNT(*) FROM packages_without_data), 0)
)
SELECT name
FROM (
	SELECT id, name, 0 AS selection_group
	FROM packages_without_data
	UNION ALL
	SELECT id, name, 1 AS selection_group
	FROM packages_with_old_data
) AS selected_packages
ORDER BY selection_group, id;