-- Migration: Restore title column
ALTER TABLE dreams ADD COLUMN title TEXT DEFAULT '';
