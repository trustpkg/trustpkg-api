WITH selected AS (
	SELECT
		p.id,
		p.name,
		p.slug,
		p.ecosystem,
		COALESCE(pd.popularity, 0)::BIGINT AS popularity,
		COALESCE(pd.is_deprecated, FALSE) AS is_deprecated
	FROM packages AS p
	LEFT JOIN packages_data AS pd
		ON pd.package_id = p.id
	WHERE p.ecosystem = $5
	ORDER BY
		CASE
			WHEN $3 = 'name' AND $4 = 'asc' THEN LOWER(p.name)
		END ASC,
		CASE
			WHEN $3 = 'name' AND $4 = 'desc' THEN LOWER(p.name)
		END DESC,
		CASE
			WHEN $3 = 'popularity' AND $4 = 'asc' THEN COALESCE(pd.popularity, 0)
		END ASC,
		CASE
			WHEN $3 = 'popularity' AND $4 = 'desc' THEN COALESCE(pd.popularity, 0)
		END DESC,
		CASE
			WHEN $3 = 'ecosystem' AND $4 = 'asc' THEN p.ecosystem
		END ASC,
		CASE
			WHEN $3 = 'ecosystem' AND $4 = 'desc' THEN p.ecosystem
		END DESC,
		p.id ASC
	LIMIT $2
	OFFSET (($1 - 1) * $2)
), affected_by_vulnerability AS (
	SELECT
		a.package_id,
		a.vulnerability_id,
		jsonb_agg(
			jsonb_build_object(
				'ecosystem', s.ecosystem,
				'introduced_version', a.introduced_version,
				'fixed_version', a.fixed_version,
				'last_affected_version', a.last_affected_version
			)
			ORDER BY a.id
		) AS affected
	FROM vulnerability_affected AS a
	JOIN selected AS s
		ON s.id = a.package_id
	GROUP BY a.package_id, a.vulnerability_id, s.ecosystem
), references_by_vulnerability AS (
	SELECT
		r.vulnerability_id,
		jsonb_agg(r.url ORDER BY r.id) AS references
	FROM vulnerability_references AS r
	WHERE r.vulnerability_id IN (
		SELECT DISTINCT vulnerability_id
		FROM affected_by_vulnerability
	)
	GROUP BY r.vulnerability_id
), vulnerability_rows AS (
	SELECT
		av.package_id,
		to_char(
			date_trunc('month', COALESCE(v.modified_at, v.published_at)),
			'YYYY-MM'
		) AS month,
		jsonb_build_object(
			'id', v.id,
			'osv_id', v.osv_id,
			'cve_id', v.cve_id,
			'summary', v.summary,
			'description', v.description,
			'severity', v.severity,
			'cvss_score', v.cvss_score,
			'cvss_vector', v.cvss_vector,
			'published_at', v.published_at,
			'modified_at', v.modified_at,
			'withdrawn_at', v.withdrawn_at,
			'affected', av.affected,
			'references', COALESCE(rv.references, '[]'::jsonb)
		) AS vulnerability
	FROM affected_by_vulnerability AS av
	JOIN vulnerabilities AS v
		ON v.id = av.vulnerability_id
	LEFT JOIN references_by_vulnerability AS rv
		ON rv.vulnerability_id = v.id
	WHERE COALESCE(v.modified_at, v.published_at) IS NOT NULL
), vulnerability_by_month AS (
	SELECT
		package_id,
		month,
		COUNT(*)::INT AS vulnerability_count,
		jsonb_agg(vulnerability ORDER BY vulnerability->>'osv_id') AS vulnerabilities
	FROM vulnerability_rows
	GROUP BY package_id, month
), vulnerability_aggregates AS (
	SELECT
		package_id,
		jsonb_object_agg(month, vulnerability_count ORDER BY month) AS vulnerability_counts,
		jsonb_object_agg(month, vulnerabilities ORDER BY month) AS vulnerabilities
	FROM vulnerability_by_month
	GROUP BY package_id
)
SELECT
	s.id,
	s.name,
	s.slug,
	s.ecosystem,
	s.popularity,
	s.is_deprecated,
	COALESCE(va.vulnerability_counts, '{}'::jsonb) AS vulnerability_counts,
	COALESCE(va.vulnerabilities, '{}'::jsonb) AS vulnerabilities
FROM selected AS s
LEFT JOIN vulnerability_aggregates AS va
	ON va.package_id = s.id
ORDER BY
	CASE
		WHEN $3 = 'name' AND $4 = 'asc' THEN LOWER(s.name)
	END ASC,
	CASE
		WHEN $3 = 'name' AND $4 = 'desc' THEN LOWER(s.name)
	END DESC,
	CASE
		WHEN $3 = 'popularity' AND $4 = 'asc' THEN s.popularity
	END ASC,
	CASE
		WHEN $3 = 'popularity' AND $4 = 'desc' THEN s.popularity
	END DESC,
	CASE
		WHEN $3 = 'ecosystem' AND $4 = 'asc' THEN s.ecosystem
	END ASC,
	CASE
		WHEN $3 = 'ecosystem' AND $4 = 'desc' THEN s.ecosystem
	END DESC,
	s.id ASC;
