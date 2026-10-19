LEFT JOIN LATERAL (
	SELECT
		jsonb_object_agg(grouped.month, grouped.vulnerability_count ORDER BY grouped.month) AS vulnerability_counts,
		jsonb_object_agg(grouped.month, grouped.vulnerabilities ORDER BY grouped.month) AS vulnerabilities
	FROM (
		SELECT
			to_char(date_trunc('month', vul.event_at), 'YYYY-MM') AS month,
			COUNT(*)::INT AS vulnerability_count,
			jsonb_agg(
				jsonb_build_object(
					'id', vul.id,
					'osv_id', vul.osv_id,
					'cve_id', vul.cve_id,
					'summary', vul.summary,
					'description', vul.description,
					'severity', vul.severity,
					'cvss_score', vul.cvss_score,
					'cvss_vector', vul.cvss_vector,
					'published_at', vul.published_at,
					'modified_at', vul.modified_at,
					'withdrawn_at', vul.withdrawn_at,
					'affected', vul.affected,
					'references', /*REFERENCES*/
				)
				ORDER BY vul.osv_id
			) AS vulnerabilities
		FROM vul
		GROUP BY date_trunc('month', vul.event_at)
	) AS grouped
) AS details ON TRUE
