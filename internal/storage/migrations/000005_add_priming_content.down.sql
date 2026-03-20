-- Migration: Remove priming content table
DROP INDEX IF EXISTS idx_priming_content_category;
DROP INDEX IF EXISTS idx_priming_content_source;
DROP TABLE IF EXISTS priming_content;
