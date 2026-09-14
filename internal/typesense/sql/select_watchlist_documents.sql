WITH selected AS (
    SELECT p.id, p.name, p.slug, p.ecosystem, d.popularity, d.is_deprecated
    FROM packages p
    JOIN packages_data d ON d.package_id = p.id
    WHERE p.ecosystem = 'npm'
      AND p.name = ANY($1::text[])
      AND array_position($1::text[], p.name) > $2
    ORDER BY array_position($1::text[], p.name)
    LIMIT $3
), vulnerability_counts AS (
    SELECT a.package_id,
           date_trunc('month', COALESCE(v.published_at, v.modified_at)) AS vulnerability_month,
           count(DISTINCT v.id)::int AS vulnerability_count
    FROM vulnerability_affected a
    JOIN vulnerabilities v ON v.id = a.vulnerability_id
    JOIN selected s ON s.id = a.package_id
    GROUP BY a.package_id, date_trunc('month', COALESCE(v.published_at, v.modified_at))
), vulnerability_json AS (
    SELECT package_id,
           jsonb_object_agg(to_char(vulnerability_month, 'YYYY-MM'), vulnerability_count) AS vulnerabilities
    FROM vulnerability_counts
    GROUP BY package_id
)
SELECT s.id::text, s.name, s.slug, s.ecosystem, s.popularity, s.is_deprecated,
    COALESCE(vj.vulnerabilities, '{}'::jsonb)
FROM selected s
LEFT JOIN vulnerability_json vj ON vj.package_id = s.id
ORDER BY array_position($1::text[], s.name);
