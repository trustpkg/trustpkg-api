WITH selected AS (
    SELECT p.id, p.name, p.slug, p.ecosystem, d.popularity, d.is_deprecated
    FROM packages p
    JOIN packages_data d ON d.package_id = p.id
    WHERE p.ecosystem = 'npm'
            AND (p.id > $1 OR d.need_update)
    ORDER BY p.id
    LIMIT $2
), vulnerability_counts AS (
    SELECT a.package_id,
           to_char(date_trunc('month', COALESCE(v.modified_at, v.published_at)), 'YYYY-MM') AS month,
           count(DISTINCT v.id)::int AS vulnerability_count
    FROM vulnerability_affected a
    JOIN vulnerabilities v ON v.id = a.vulnerability_id
    JOIN selected s ON s.id = a.package_id
    WHERE COALESCE(v.modified_at, v.published_at) >= date_trunc('month', CURRENT_DATE) - INTERVAL '11 months'
    GROUP BY a.package_id, date_trunc('month', COALESCE(v.modified_at, v.published_at))
), vulnerability_json AS (
    SELECT package_id, jsonb_object_agg(month, vulnerability_count) AS vulnerabilities
    FROM vulnerability_counts
    GROUP BY package_id
)
SELECT s.id::text, s.name, s.slug, s.ecosystem, s.popularity, s.is_deprecated,
       COALESCE(vj.vulnerabilities, '{}'::jsonb)
FROM selected s
LEFT JOIN vulnerability_json vj ON vj.package_id = s.id
ORDER BY s.id;
