-- name: CreateDelivery :one
INSERT INTO deliveries (webhook_id, event_id)
VALUES ($1, $2)
RETURNING *;
-- name: UpdateDeliveryStatus :one
UPDATE deliveries
SET status = ($2)
WHERE id = $1
RETURNING *;