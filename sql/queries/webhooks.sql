-- name: ListWebhooks :many
SELECT id,
    event_name,
    target_url
FROM webhooks;
-- name: CreateWebhook :exec
INSERT INTO webhooks (event_name, target_url)
VALUES ($1, $2);