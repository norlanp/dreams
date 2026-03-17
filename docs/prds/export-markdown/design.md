# Export to Markdown

**Status:** Brainstormed
**Date:** 2026-03-16
**Participants:** User, AI

## Problem Statement

Users need a way to backup their dream journal data in a portable, human-readable format. Markdown files with date-time prefixed filenames provide an easy-to-read backup that can be version controlled, shared, or archived.

## Goals

- Export all dreams to individual Markdown files for backup
- Support both CLI flag and TUI menu option
- Use date-time prefixed filenames for easy sorting
- Include minimal frontmatter with date
- Create export directory if it doesn't exist
- Overwrite existing files (idempotent operation)
- Show progress/confirmation in TUI

## Approach

**Selected:** Go-native export function

Pure Go implementation using existing repository pattern. Reads all dreams from SQLite and writes individual Markdown files. No external dependencies needed.

### Alternatives Considered

- **Python script:** Overkill for simple file I/O, adds unnecessary complexity
- **SQLite `.mode markdown`:** Can't control filename format, limited customization

## Architecture

### Components

```
cmd/main.go              # Add --export flag parsing
internal/
  export/
    exporter.go         # Export function
    exporter_test.go    # Unit tests
  tui/
    update.go           # Add 'e' keybinding handler
    views.go            # Add export confirmation view
```

### Data Flow

1. **CLI Flow:**
   - Parse `--export <directory>` flag
   - Call `export.ExportAll(repo, directory)`
   - Print success/failure to stdout
   - Exit

2. **TUI Flow:**
   - User presses 'e' in list view
   - Prompt for export directory (default: `./dreams-export`)
   - Show confirmation: "Export N dreams to <dir>?"
   - On confirm: run export in background with loading state
   - Show success message with file count

### Export Format

**Filename:** `{timestamp}-dream.md`
- Format: `2006-01-02-15-04-05-dream.md`
- Example: `2026-03-15-08-30-00-dream.md`

**File content:**
```markdown
---
date: 2026-03-15
---

Dream content here...
Multiple lines preserved.
```

## Error Handling

| Error | Handling |
|-------|----------|
| Directory not writable | Clear error message: "Cannot write to directory: <path>" |
| Zero dreams | "No dreams to export" |
| Directory creation fails | Wrap and return error |
| Partial write failure | Continue with remaining dreams, report count at end |

## Testing Strategy

**Unit tests (`exporter_test.go`):**
- Export function with mock repository
- Filename generation for various dates
- Directory creation
- Frontmatter formatting

**Integration tests:**
- Export actual dreams from test database
- Verify files created with correct content
- Verify idempotent behavior (re-run safe)

**Edge cases:**
- Empty database
- Special characters in dream content
- Very long content
- Directory already exists
- Directory not writable

**Test naming:** `TestExport_ShouldCreateFileWithFrontmatter`

## Implementation Notes

- Export runs synchronously (dreams are small, operation is fast)
- Use `os.MkdirAll` for directory creation
- Use `filepath.Join` for cross-platform paths
- Escape special characters in content if needed
- TUI shows loading spinner during export
- Success message: "Exported N dreams to <directory>"

## Open Questions

- Export metadata (analysis clusters)? **Decision:** No - keep it simple
- Support export filtering in future? **Decision:** Can add later
- Resume capability needed? **Decision:** No - fresh export each time
