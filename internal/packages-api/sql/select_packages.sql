WITH selected AS MATERIALIZED (
	SELECT
		p.id,
		p.name,
		p.slug,
		p.ecosystem,
		COALESCE(pd.popularity, 0)::BIGINT AS popularity,
		COALESCE(pd.is_deprecated, FALSE) AS is_deprecated
	FROM /*PACKAGE_SOURCE*/
	WHERE /*PACKAGE_FILTER*/
	ORDER BY /*SELECT_ORDER*/
	LIMIT $2
	OFFSET (($1 - 1) * $2)
)
SELECT
	s.id,
	s.name,
	s.slug,
	s.ecosystem,
	s.popularity,
	s.is_deprecated
	/*CURRENT_STATUS_COLUMN*/
	/*VULNERABILITY_COLUMNS*/
FROM selected AS s
/*VULNERABILITY_DETAILS*/
/*CURRENT_STATUS_DETAILS*/
ORDER BY /*RESULT_ORDER*/;
