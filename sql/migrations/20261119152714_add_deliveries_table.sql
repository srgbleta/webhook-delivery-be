-- +goose Up
CREATE TABLE deliveries (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    webhook_id BIGINT NOT NULL REFERENCES webhooks(id),
    event_id BIGINT NOT NULL REFERENCES events(id),
    status TEXT NOT NULL DEFAULT 'in_progress' CONSTRAINT deliveries_status_check CHECK (status IN ('in_progress', 'done', 'failed'))
);
-- +goose Down
DROP TABLE deliveries;