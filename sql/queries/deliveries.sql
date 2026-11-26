-- name: CreateDelivery :one
INSERT INTO deliveries (webhook_id, event_id)
VALUES ($1, $2)
RETURNING *;
-- name: UpdateDeliveryStatus :one
UPDATE deliveries
SET status = ($2)
WHERE id = $1
RETURNING *;
-- name: ListPendingDeliveries :many
SELECT *
FROM deliveries
WHERE status = 'pending';
-- name: ListPendingDeliveriesWithWebhooks :many
SELECT d.id AS delivery_id,
    d.event_id AS event_id,
    d.status AS delivery_status,
    w.id AS webhook_id,
    w.event_name AS event_name,
    w.target_url AS target_url,
    e.data AS event_data
FROM deliveries AS d
    JOIN webhooks AS w ON d.webhook_id = w.id
    JOIN events AS e ON d.event_id = e.id
WHERE d.status = 'pending';