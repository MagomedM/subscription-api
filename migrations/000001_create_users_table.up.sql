CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL DEFAULT 0 CHECK (price >= 0),
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
);

CREATE INDEX idx_users_start_date ON users(start_date);
CREATE INDEX idx_users_end_date ON users(end_date);
CREATE INDEX idx_users_service_name ON users(service_name);
CREATE INDEX idx_users_user_id ON users(user_id);