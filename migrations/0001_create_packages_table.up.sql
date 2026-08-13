CREATE TABLE IF NOT EXISTS packages (
    id BIGSERIAL PRIMARY KEY UNIQUE,
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT,
    license TEXT,
    author TEXT,
    repository TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NULL,
    maintainers TEXT[],
    keywords TEXT[],
    deprecated BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_packages_name ON packages(LOWER(name));