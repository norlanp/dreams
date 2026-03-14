-- name: CreateDream :one
INSERT INTO dreams (title, content, created_at, updated_at)
VALUES (?, ?, ?, ?)
RETURNING id, title, content, created_at, updated_at;

-- name: ListDreams :many
SELECT id, title, content, created_at, updated_at
FROM dreams
ORDER BY created_at DESC;

-- name: GetDream :one
SELECT id, title, content, created_at, updated_at
FROM dreams
WHERE id = ?;

-- name: UpdateDream :one
UPDATE dreams
SET title = ?, content = ?, updated_at = ?
WHERE id = ?
RETURNING id, title, content, created_at, updated_at;

-- name: DeleteDream :exec
DELETE FROM dreams WHERE id = ?;
