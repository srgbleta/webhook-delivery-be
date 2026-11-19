-- name: CreateEvent :one
INSERT INTO events (name, data)
VALUES ($1, $2)
RETURNING *;