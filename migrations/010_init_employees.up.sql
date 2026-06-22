CREATE TABLE IF NOT EXISTS employees (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    phone_number VARCHAR(20),
    private_email VARCHAR(50),
    street VARCHAR(100),
    country BIGINT NOT NULL REFERENCES countries(country_id),
    city VARCHAR(100),
    date_of_birth DATE,
    hire_date DATE NOT NULL DEFAULT CURRENT_DATE,
    position_id BIGINT REFERENCES positions(id),
    created_at TIMESTAMP DEFAULT now()
);
