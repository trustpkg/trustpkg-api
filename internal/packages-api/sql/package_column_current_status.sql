,
	COALESCE((
		SELECT jsonb_agg(
			jsonb_build_object(
				'code', CASE WHEN status.has_unfixed THEN 'vulnerable_no_fix' ELSE 'vulnerable_fix_available' END,
				'label', CASE WHEN status.has_unfixed THEN 'Vulnerable, no fix available' ELSE 'Vulnerable, fix available' END,
				'osv_id', status.osv_id,
				'cve_id', status.cve_id,
				'severity', status.severity,
				'affected', status.affected,
				'fixed_versions', status.fixed_versions
			)
			ORDER BY status.osv_id
		)
		FROM (
			SELECT
				v.osv_id,
				v.cve_id,
				v.severity,
				current_aff.has_unfixed,
				current_aff.affected,
				current_aff.fixed_versions
			FROM (
				SELECT
					a.vulnerability_id,
					BOOL_OR(NULLIF(BTRIM(a.fixed_version), '') IS NULL) AS has_unfixed,
					jsonb_agg(
						jsonb_build_object(
							'ecosystem', pkg.ecosystem,
							'introduced_version', a.introduced_version,
							'fixed_version', a.fixed_version,
							'last_affected_version', a.last_affected_version
						)
						ORDER BY a.id
					) AS affected,
					COALESCE(
						array_agg(DISTINCT NULLIF(BTRIM(a.fixed_version), '') ORDER BY NULLIF(BTRIM(a.fixed_version), ''))
							FILTER (WHERE NULLIF(BTRIM(a.fixed_version), '') IS NOT NULL),
						ARRAY[]::TEXT[]
					) AS fixed_versions
				FROM vulnerability_affected AS a
				WHERE a.package_id = pkg.id
				GROUP BY a.vulnerability_id
			) AS current_aff
			JOIN vulnerabilities AS v ON v.id = current_aff.vulnerability_id
			WHERE v.withdrawn_at IS NULL
			  AND (
				v.published_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
				OR v.modified_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
			  )
		) AS status
	), jsonb_build_array(
		jsonb_build_object(
			'code', 'not_vulnerable',
			'label', 'No known vulnerabilities',
			'osv_id', NULL,
			'cve_id', NULL,
			'severity', NULL,
			'affected', '[]'::jsonb,
			'fixed_versions', ARRAY[]::TEXT[]
		)
	)) AS current_statuses
