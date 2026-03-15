-- name: CreateAnalysis :one
INSERT INTO dream_analysis (analysis_date, dream_count, n_clusters, results_json, created_at)
VALUES (?, ?, ?, ?, ?)
RETURNING id, analysis_date, dream_count, n_clusters, results_json, created_at;

-- name: CreateCluster :one
INSERT INTO dream_clusters (analysis_id, cluster_id, dream_count, top_terms, dream_ids, created_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, analysis_id, cluster_id, dream_count, top_terms, dream_ids, created_at;

-- name: GetLatestAnalysis :one
SELECT id, analysis_date, dream_count, n_clusters, results_json, created_at
FROM dream_analysis
ORDER BY created_at DESC
LIMIT 1;

-- name: GetAnalysisClusters :many
SELECT id, analysis_id, cluster_id, dream_count, top_terms, dream_ids, created_at
FROM dream_clusters
WHERE analysis_id = ?
ORDER BY cluster_id;

-- name: ListAnalysisHistory :many
SELECT id, analysis_date, dream_count, n_clusters, results_json, created_at
FROM dream_analysis
ORDER BY created_at DESC;

-- name: DeleteAnalysis :exec
DELETE FROM dream_analysis WHERE id = ?;
