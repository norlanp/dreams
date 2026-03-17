# Export to Markdown

## Overview

A feature to export all dreams from the SQLite database to individual Markdown files for backup and portability.

**Status:** Planning  
**Target:** v1.2.0  
**Priority:** Medium

## Problem Statement

Users need a way to backup their dream journal data in a portable, human-readable format. Markdown files with date-time prefixed filenames provide an easy-to-read backup that can be version controlled, shared, or archived.

## Goals

- Export all dreams to individual Markdown files for backup
- Support both CLI flag and TUI menu option
- Use date-time prefixed filenames for easy chronological sorting
- Include minimal frontmatter with date metadata
- Create export directory if it doesn't exist
- Overwrite existing files (idempotent operation)
- Show progress/confirmation in TUI

## Requirements

### Functional

1. **CLI Export:** Accept `--export <directory>` flag to export all dreams
2. **TUI Export:** Add 'e' keybinding in list view to trigger export
3. **Filename Format:** `{timestamp}-dream.md` (e.g., `2026-03-15-08-30-00-dream.md`)
4. **File Content:** Markdown with YAML frontmatter containing date
5. **Directory Handling:** Create directory if missing, overwrite files if present
6. **Feedback:** Display count of exported dreams and success/failure message

### Non-Functional

- **Performance:** Export should complete in < 5 seconds for 1000 dreams
- **Safety:** Atomic writes (write to temp file, then rename)
- **Cross-platform:** Windows/Unix compatible file paths
- **Testability:** 100% unit test coverage for export logic

## Architecture

### Components

```
cmd/main.go              # Add --export flag parsing
internal/
  export/
    exporter.go         # Export function
    exporter_test.go    # Unit tests
  tui/
    model.go            # Add ExportMode to PageState
    update.go           # Add 'e' keybinding handler
    views.go            # Add export confirmation and progress views
```

### Data Flow

**CLI Flow:**
1. Parse `--export <directory>` flag
2. Call `export.ExportAll(repo, directory)`
3. Print success/failure to stdout
4. Exit with appropriate code

**TUI Flow:**
1. User presses 'e' in list view
2. Switch to ExportMode with directory prompt (default: `./dreams-export`)
3. Show confirmation: "Export N dreams to <dir>?"
4. On confirm: run export with loading spinner
5. Show success message with file count

### Export Format

**Filename:** `2006-01-02-15-04-05-dream.md`

**File content:**
```markdown
---
date: 2026-03-15
---

Dream content here...
Multiple lines preserved.
```

## Interface Design

### CLI Usage

```bash
# Export to current directory
dreams --export .

# Export to specific directory
dreams --export ~/backups/dreams

# Success output
Exported 42 dreams to /home/user/backups/dreams
```

### TUI Flow

```
[List View]
  Dreams (3)  [e] Export  [/] Search  [q] Quit

  > 2026-03-15 Morning flight
    2026-03-14 Underwater city
    2026-03-13 Lost in school

[Press 'e' -> Export Mode]
  Export Dreams
  Directory: ./dreams-export
  [Enter] Confirm  [Esc] Cancel

[After Confirm -> Loading]
  Exporting dreams...
  [spinner]

[After Complete -> Success]
  Exported 3 dreams to ./dreams-export
  [Enter] Return to list  [q] Quit
```

## Error Handling

| Error | Handling |
|-------|----------|
| Directory not writable | "Cannot write to directory: <path>" |
| Zero dreams | "No dreams to export" |
| Directory creation fails | Wrap and return error with context |
| Partial write failure | Continue with remaining, report count at end |
| Invalid directory path | Validate early, clear error message |

## Testing Strategy

### Unit Tests (`exporter_test.go`)

- `TestExport_ShouldCreateFileWithFrontmatter` - Verify file format
- `TestExport_ShouldGenerateCorrectFilename` - Timestamp format
- `TestExport_ShouldCreateDirectory` - Directory creation
- `TestExport_ShouldOverwriteExistingFiles` - Idempotent behavior
- `TestExport_ShouldHandleSpecialCharacters` - Content escaping

### Integration Tests

- Export actual dreams from test database
- Verify files created with correct content
- Verify idempotent behavior (re-run safe)
- Verify cross-platform path handling

### Edge Cases

- Empty database
- Special characters in dream content (`*`, `_`, `#`, etc.)
- Very long content (10KB+)
- Directory already exists with existing files
- Directory not writable
- Interrupted export (partial files)

## Security Considerations

- Validate directory path (no traversal attacks)
- Limit filename length
- Sanitize content (Markdown should escape special chars naturally)

## Open Questions (Resolved)

- Export metadata (analysis clusters)? **Decision:** No - keep it simple
- Support export filtering in future? **Decision:** Can add later
- Resume capability needed? **Decision:** No - fresh export each time
- Include tags/categories in frontmatter? **Decision:** No - minimal frontmatter for now

## Dependencies

- `os`, `path/filepath` - Standard library only
- Existing `internal/storage` repository interface

## Success Metrics

- All unit tests passing
- Export 100 dreams in < 1 second
- Zero data loss during export
- Successfully handles special characters in content

## Future Considerations

- Export with tags/categories when those features are added
- Selective export (date range, search results)
- Export to other formats (JSON, CSV)
- Compressed export (tar.gz, zip)
