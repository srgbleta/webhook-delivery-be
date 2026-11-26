-- +goose Up
ALTER TABLE deliveries DROP CONSTRAINT deliveries_status_check;
ALTER TABLE deliveries
ADD CONSTRAINT deliveries_status_check CHECK (
        status IN ('pending', 'in_progress', 'done', 'failed')
    );
ALTER TABLE deliveries
ALTER COLUMN status
SET DEFAULT 'pending';
-- +goose Down
ALTER TABLE deliveries DROP CONSTRAINT deliveries_status_check;
ALTER TABLE deliveries
ADD CONSTRAINT deliveries_status_check CHECK (
        status IN ('in_progress', 'done', 'failed')
    );
ALTER TABLE deliveries
ALTER COLUMN status
SET DEFAULT 'in_progress';