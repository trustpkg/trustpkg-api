LEFT JOIN LATERAL (
	SELECT
		jsonb_object_agg(grouped.month, grouped.vulnerability_count ORDER BY grouped.month) AS vulnerability_counts,
		jsonb_object_agg(grouped.month, grouped.vulnerabilities ORDER BY grouped.month) AS vulnerabilities
	FROM (
		SELECT
			to_char(date_trunc('month', COALESCE(v.published_at, v.modified_at)), 'YYYY-MM') AS month,
			COUNT(*)::INT AS vulnerability_count,
			jsonb_agg(
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
					'affected', COALESCE((
						SELECT jsonb_agg(
							jsonb_build_object(
								'ecosystem', s.ecosystem,
								'introduced_version', affected.introduced_version,
								'fixed_version', affected.fixed_version,
								'last_affected_version', affected.last_affected_version
							)
							ORDER BY affected.id
						)
						FROM vulnerability_affected AS affected
						WHERE affected.package_id = s.id
						  AND affected.vulnerability_id = v.id
					), '[]'::jsonb),
					'references', COALESCE((
						SELECT jsonb_agg(reference.url ORDER BY reference.id)
						FROM vulnerability_references AS reference
						WHERE reference.vulnerability_id = v.id
					), '[]'::jsonb)
				)
				ORDER BY v.osv_id
			) AS vulnerabilities
		FROM vulnerabilities AS v
		JOIN (
			SELECT DISTINCT package_affected.vulnerability_id
			FROM vulnerability_affected AS package_affected
			WHERE package_affected.package_id = s.id
		) AS package_vulnerabilities
			ON package_vulnerabilities.vulnerability_id = v.id
		WHERE $4::int = 0
		   OR COALESCE(v.published_at, v.modified_at) >=
		      date_trunc('month', CURRENT_DATE) - (($4::int - 1) * INTERVAL '1 month')
		GROUP BY date_trunc('month', COALESCE(v.published_at, v.modified_at))
	) AS grouped
) AS details ON TRUE
