CREATE TYPE orders_status AS ENUM ('NEW', 'REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
    number BIGINT NOT NULL,
    user_id INTEGER NOT NULL,
    status orders_status NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(10, 2),
    uploaded_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_orders_number ON orders(number);
CREATE INDEX idx_orders_user_id ON orders(user_id);
