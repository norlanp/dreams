-- name: CreateDream :one
INSERT INTO dreams (content, created_at, updated_at)
VALUES (?, ?, ?)
RETURNING id, content, created_at, updated_at;

-- name: ListDreams :many
SELECT id, content, created_at, updated_at
FROM dreams
ORDER BY created_at ASC;

-- name: GetDream :one
SELECT id, content, created_at, updated_at
FROM dreams
WHERE id = ?;

-- name: UpdateDream :one
UPDATE dreams
SET content = ?, updated_at = ?
WHERE id = ?
RETURNING id, content, created_at, updated_at;

-- name: DeleteDream :exec
DELETE FROM dreams WHERE id = ?;
