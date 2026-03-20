-- name: GetPrimingCache :one
SELECT id, source, payload_json, fetched_at, updated_at
FROM priming_cache
WHERE source = ?;

-- name: UpsertPrimingCache :exec
INSERT INTO priming_cache (source, payload_json, fetched_at, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(source) DO UPDATE SET
  payload_json = excluded.payload_json,
  fetched_at = excluded.fetched_at,
  updated_at = excluded.updated_at;

-- name: InsertPrimingLog :exec
INSERT INTO priming_logs (created_at, source, outcome, detail, content)
VALUES (?, ?, ?, ?, ?);

-- name: ListPrimingLogs :many
SELECT id, created_at, source, outcome, detail, content
FROM priming_logs
ORDER BY created_at DESC;

-- name: InsertPrimingContent :exec
INSERT INTO priming_content (source, title, content, category, url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListPrimingContent :many
SELECT id, source, title, content, category, url, created_at, updated_at
FROM priming_content
ORDER BY created_at DESC;

-- name: GetPrimingContentBySource :many
SELECT id, source, title, content, category, url, created_at, updated_at
FROM priming_content
WHERE source = ?;

-- name: GetPrimingContentByCategory :many
SELECT id, source, title, content, category, url, created_at, updated_at
FROM priming_content
WHERE category = ?;

-- name: DeletePrimingContent :exec
DELETE FROM priming_content WHERE id = ?;

-- name: CountPrimingContent :one
SELECT COUNT(*) FROM priming_content;
