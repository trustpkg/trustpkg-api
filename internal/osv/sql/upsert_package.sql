INSERT INTO packages (name, ecosystem)
VALUES ($1, $2)
ON CONFLICT (ecosystem, name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;
