DELETE FROM packages 
WHERE name = ANY($1::text[]);