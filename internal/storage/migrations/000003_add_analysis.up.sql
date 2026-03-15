-- Migration: Add dream analysis tables
CREATE TABLE IF NOT EXISTS dream_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    analysis_date TEXT NOT NULL,
    dream_count INTEGER NOT NULL,
    n_clusters INTEGER NOT NULL,
    results_json TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dream_clusters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    analysis_id INTEGER NOT NULL,
    cluster_id INTEGER NOT NULL,
    dream_count INTEGER NOT NULL,
    top_terms TEXT NOT NULL,
    dream_ids TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (analysis_id) REFERENCES dream_analysis(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_dream_clusters_analysis_id ON dream_clusters(analysis_id);
