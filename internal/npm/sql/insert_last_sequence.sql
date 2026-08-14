INSERT INTO npm_replication_state (id, last_seq)
VALUES (1, $1)
ON CONFLICT (id)
DO UPDATE SET last_seq = EXCLUDED.last_seq;