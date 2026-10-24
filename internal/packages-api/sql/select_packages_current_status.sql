LEFT JOIN LATERAL (
	SELECT COALESCE(
		jsonb_agg(
			jsonb_build_object(
				'code', CASE
					WHEN status.has_unfixed_range THEN 'vulnerable_no_fix'
					ELSE 'vulnerable_fix_available'
				END,
				'label', CASE
					WHEN status.has_unfixed_range THEN 'Vulnerable, no fix available'
					ELSE 'Vulnerable, fix available'
				END,
				'osv_id', status.osv_id,
				'cve_id', status.cve_id,
				'severity', status.severity,
				'affected', status.affected,
				'fixed_versions', status.fixed_versions
			)
			ORDER BY status.osv_id
		),
		jsonb_build_array(
			jsonb_build_object(
				'code', 'not_vulnerable',
				'label', 'No known vulnerabilities',
				'osv_id', NULL,
				'cve_id', NULL,
				'severity', NULL,
				'affected', '[]'::jsonb,
				'fixed_versions', ARRAY[]::TEXT[]
			)
		)
	) AS current_statuses
	FROM (
		SELECT
			v.osv_id,
			v.cve_id,
			v.severity,
			BOOL_OR(NULLIF(BTRIM(affected.fixed_version), '') IS NULL) AS has_unfixed_range,
			COALESCE(
				jsonb_agg(
					jsonb_build_object(
						'ecosystem', s.ecosystem,
						'introduced_version', affected.introduced_version,
						'fixed_version', affected.fixed_version,
						'last_affected_version', affected.last_affected_version
					)
					ORDER BY affected.id
				),
				'[]'::jsonb
			) AS affected,
			COALESCE(
				array_agg(DISTINCT NULLIF(BTRIM(affected.fixed_version), '') ORDER BY NULLIF(BTRIM(affected.fixed_version), ''))
					FILTER (WHERE NULLIF(BTRIM(affected.fixed_version), '') IS NOT NULL),
				ARRAY[]::TEXT[]
			) AS fixed_versions
		FROM vulnerability_affected AS affected
		JOIN vulnerabilities AS v ON v.id = affected.vulnerability_id
		WHERE affected.package_id = s.id
		  AND v.withdrawn_at IS NULL
		  AND (
			v.published_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
			OR v.modified_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
		  )
		GROUP BY v.id, v.osv_id, v.cve_id, v.severity
	) AS status
) AS current_status ON TRUE