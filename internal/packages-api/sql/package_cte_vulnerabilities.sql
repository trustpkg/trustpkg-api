, aff AS MATERIALIZED (
	SELECT
		a.vulnerability_id,
		jsonb_agg(
			jsonb_build_object(
				'ecosystem', pkg.ecosystem,
				'introduced_version', a.introduced_version,
				'fixed_version', a.fixed_version,
				'last_affected_version', a.last_affected_version
			)
			ORDER BY a.id
		) AS affected,
		BOOL_OR(NULLIF(BTRIM(a.fixed_version), '') IS NULL) AS has_unfixed
	FROM pkg
	JOIN vulnerability_affected AS a ON a.package_id = pkg.id
	GROUP BY a.vulnerability_id
), vul AS MATERIALIZED (
	SELECT
		v.id,
		v.osv_id,
		v.cve_id,
		v.summary,
		/*DESCRIPTION_COLUMN*/ AS description,
		v.severity,
		v.cvss_score,
		v.cvss_vector,
		v.published_at,
		v.modified_at,
		v.withdrawn_at,
		COALESCE(v.published_at, v.modified_at) AS event_at,
		aff.affected,
		aff.has_unfixed
	FROM aff
	JOIN vulnerabilities AS v ON v.id = aff.vulnerability_id
	WHERE TRUE
	/*VULNERABILITY_FILTERS*/
)
