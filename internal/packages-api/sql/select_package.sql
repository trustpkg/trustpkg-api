WITH pkg AS MATERIALIZED (
	SELECT
		p.id,
		p.name,
		p.slug,
		p.ecosystem,
		COALESCE(pd.popularity, 0)::BIGINT AS popularity,
		COALESCE(pd.is_deprecated, FALSE) AS is_deprecated
	FROM packages AS p
	LEFT JOIN packages_data AS pd ON pd.package_id = p.id
	WHERE p.ecosystem = /*ECOSYSTEM*/
	  AND (p.slug = /*IDENTIFIER*/ OR p.name = /*IDENTIFIER*/)
	ORDER BY (p.name = /*IDENTIFIER*/) DESC, p.id ASC
	LIMIT 1
)
/*CTES*/
SELECT
	pkg.id,
	pkg.name,
	pkg.slug,
	pkg.ecosystem,
	pkg.popularity,
	pkg.is_deprecated
	/*COLUMNS*/
FROM pkg
/*JOINS*/;
