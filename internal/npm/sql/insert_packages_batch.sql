INSERT INTO packages (name) 
SELECT unnest($1::text[]) 
ON CONFLICT (name) DO NOTHING;