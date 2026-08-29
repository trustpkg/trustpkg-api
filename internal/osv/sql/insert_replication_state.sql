INSERT INTO osv_replication_state (ecosystem)
VALUES ($1)
ON CONFLICT (ecosystem) DO NOTHING;