INSERT INTO vulnerability_references (vulnerability_id, url)
VALUES ($1, $2)
ON CONFLICT (vulnerability_id, url) DO NOTHING;
