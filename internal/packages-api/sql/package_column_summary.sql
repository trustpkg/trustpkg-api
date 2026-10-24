,
	(
		SELECT jsonb_build_object(
			'total', COUNT(*),
			'unfixed', COUNT(*) FILTER (WHERE has_unfixed),
			'max_cvss_score', MAX(cvss_score),
			'latest_published_at', MAX(event_at),
			'by_severity', COALESCE((
				SELECT jsonb_object_agg(grouped.severity, grouped.total)
				FROM (
					SELECT COALESCE(UPPER(severity), 'UNKNOWN') AS severity, COUNT(*) AS total
					FROM vul
					GROUP BY 1
				) AS grouped
			), '{}'::jsonb)
		)
		FROM vul
	) AS summary
