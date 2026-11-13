-- +goose Up
CREATE TABLE webhooks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_name VARCHAR(256) NOT NULL,
    target_url TEXT NOT NULL
);
-- +goose Down
DROP TABLE webhooks;