# Export to Markdown - Acceptance Checklist

**Feature:** Export dreams to Markdown files  
**Status:** Draft → Ready for Development  
**Date:** 2026-03-16

## Functional Requirements

### CLI Export
- [ ] `dreams --export .` exports to current directory
- [ ] `dreams --export /path/to/dir` exports to specified directory
- [ ] Creates directory if it doesn't exist
- [ ] Overwrites existing files without error
- [ ] Prints "Exported N dreams to <path>" on success
- [ ] Prints clear error message on failure
- [ ] Exits with code 0 on success, non-zero on error

### TUI Export
- [ ] Pressing 'e' in list view enters export mode
- [ ] Shows directory prompt with default "./dreams-export"
- [ ] User can type custom directory path
- [ ] Shows confirmation: "Export N dreams to <dir>?"
- [ ] [Enter] confirms, [Esc] cancels
- [ ] Shows loading spinner during export
- [ ] Shows success message with file count
- [ ] [Enter] returns to list view
- [ ] 'q' quits from success view

### File Format
- [ ] Filename format: `2006-01-02-15-04-05-dream.md`
- [ ] File contains YAML frontmatter with date
- [ ] Frontmatter format:
  ```markdown
  ---
  date: 2026-03-15
  ---
  ```
- [ ] Dream content follows frontmatter
- [ ] Multi-line content preserved correctly
- [ ] Special characters handled properly

### Data Integrity
- [ ] All dreams exported (none missing)
- [ ] Dream content matches database exactly
- [ ] Timestamps in filename match dream creation time
- [ ] Date in frontmatter matches dream date

## Error Handling

### CLI Errors
- [ ] Empty database: "No dreams to export"
- [ ] Non-writable directory: "Cannot write to directory: <path>"
- [ ] Invalid path: Clear error message
- [ ] Permission denied: Appropriate error

### TUI Errors
- [ ] Empty database: Shows "No dreams to export" message
- [ ] Write error: Shows error with option to retry or cancel
- [ ] Directory creation fails: Clear error message

## Performance

- [ ] Export 100 dreams in < 1 second
- [ ] Export 1000 dreams in < 5 seconds
- [ ] No UI freezing during TUI export (async with spinner)

## Code Quality

- [ ] Unit tests for exporter.go (100% coverage)
- [ ] Tests for filename generation
- [ ] Tests for frontmatter generation
- [ ] Tests for directory creation
- [ ] Tests for file overwriting
- [ ] Tests for special character handling
- [ ] All tests pass
- [ ] `go vet` clean
- [ ] `gofmt` formatted
- [ ] No linting errors

## Cross-Platform

- [ ] Works on macOS
- [ ] Works on Linux
- [ ] Works on Windows
- [ ] Handles path separators correctly
- [ ] No hardcoded path separators

## Edge Cases

- [ ] Empty database handled gracefully
- [ ] Single dream export works
- [ ] Very long content (10KB+) handled
- [ ] Special characters: `*`, `_`, `#`, `>`, `<`, `|`
- [ ] Unicode content preserved
- [ ] Newlines in content preserved
- [ ] Directory with existing files works
- [ ] Nested directory creation works
- [ ] Interrupted export (partial files handled)

## Completion Criteria

- [x] PRD created and reviewed
- [x] Design document finalized
- [x] Execution plan created
- [x] Acceptance checklist created
- [x] Todos generated
- [ ] All acceptance criteria met
- [ ] Tests passing
- [ ] Code review approved
- [ ] Documentation updated

## Sign-Off

- [ ] Developer review complete
- [ ] Code review approved
- [ ] QA testing complete
- [ ] Documentation updated
- [ ] Ready for merge

---

**Notes:**
- Atomic writes required (temp file + rename)
- No partial file corruption on interrupt
- Idempotent operation (re-running produces same result)
