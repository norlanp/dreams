-- Migration: Add night priming cache and logs
CREATE TABLE IF NOT EXISTS priming_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL UNIQUE,
    payload_json TEXT NOT NULL,
    fetched_at DATETIME NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS priming_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL,
    source TEXT NOT NULL,
    outcome TEXT NOT NULL,
    detail TEXT NOT NULL,
    content TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_priming_logs_created_at ON priming_logs(created_at DESC);
