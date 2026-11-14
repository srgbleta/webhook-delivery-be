-- name: ListWebhooks :many
SELECT id,
    event_name,
    target_url
FROM webhooks;
-- name: CreateWebhook :one
INSERT INTO webhooks (event_name, target_url)
VALUES ($1, $2)
RETURNING *;