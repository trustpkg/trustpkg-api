,
	COALESCE(details.vulnerability_counts, '{}'::jsonb) AS vulnerability_counts,
	COALESCE(details.vulnerabilities, '{}'::jsonb) AS vulnerabilities
