CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_active boolean default TRUE,
    created_at TIMESTAMP DEFAULT now(),
    created_by BIGINT NULL REFERENCES users(id),
    updated_at TIMESTAMP,
    updated_by BIGINT NULL REFERENCES users(id)
);
