CREATE TABLE systems (
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    needs_repair BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
