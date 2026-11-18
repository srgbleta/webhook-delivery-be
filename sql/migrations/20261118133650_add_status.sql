-- +goose Up
ALTER TABLE webhooks
ADD COLUMN status TEXT NOT NULL DEFAULT 'active' CONSTRAINT webhooks_status_check CHECK (status IN ('active', 'paused', 'disabled'));
-- +goose Down
ALTER TABLE webhooks DROP COLUMN status;