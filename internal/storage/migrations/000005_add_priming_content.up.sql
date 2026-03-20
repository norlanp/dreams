-- Migration: Add priming content table for community resources
CREATE TABLE IF NOT EXISTS priming_content (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category TEXT,
    url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_priming_content_source ON priming_content(source);
CREATE INDEX IF NOT EXISTS idx_priming_content_category ON priming_content(category);
