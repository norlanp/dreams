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
