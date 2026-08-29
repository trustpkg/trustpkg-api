INSERT INTO packages (name, ecosystem)
SELECT unnest($1::text[]), 'npm'
ON CONFLICT (ecosystem, name) DO NOTHING;