CREATE TABLE withdrawn_users (
    order_number BIGINT NOT NULL,
    user_id INTEGER NOT NULL,
    withdrawn NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_withdrawn_users_user_id_order_number ON withdrawn_users(user_id, order_number);
