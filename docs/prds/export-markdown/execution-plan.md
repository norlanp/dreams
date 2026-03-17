# Export to Markdown - Execution Plan

**Status:** Planning → In-Progress
**Target:** v1.2.0

## Phase 1: Foundation
**Goal:** Create export package with core functionality

### Tasks
1. Create `internal/export/` directory
2. Create `exporter.go` with ExportAll function
3. Implement filename generation (timestamp format)
4. Implement frontmatter generation
5. Implement file writing with atomic rename
6. Implement directory creation

### Deliverables
- `internal/export/exporter.go` with ExportAll function
- Unit tests for filename generation
- Unit tests for frontmatter formatting

### Definition of Done
- [ ] ExportAll function compiles
- [ ] Unit tests pass
- [ ] Handles empty database gracefully

## Phase 2: CLI Integration
**Goal:** Add --export flag to main.go

### Tasks
1. Add --export flag parsing in cmd/main.go
2. Wire export function to flag handler
3. Add stdout output for success/failure
4. Set appropriate exit codes

### Deliverables
- Updated cmd/main.go with --export support
- Help text updated

### Definition of Done
- [ ] `dreams --export <dir>` works
- [ ] Exit code 0 on success, non-zero on failure
- [ ] Clear error messages to stdout

## Phase 3: TUI Integration
**Goal:** Add export UI flow

### Tasks
1. Add ExportMode to PageState enum in tui/model.go
2. Add exportDirectory field to model
3. Add 'e' keybinding handler in list view
4. Create export prompt view (directory input)
5. Create export confirmation view
6. Create export loading view with spinner
7. Create export success/failure view

### Deliverables
- Updated tui/model.go
- Updated tui/update.go with handlers
- Updated tui/views.go with new views

### Definition of Done
- [ ] 'e' key triggers export flow
- [ ] Directory defaults to ./dreams-export
- [ ] Shows confirmation before export
- [ ] Shows spinner during export
- [ ] Shows success message with count

## Phase 4: Integration Testing
**Goal:** End-to-end testing of both CLI and TUI

### Tasks
1. Test CLI export with actual database
2. Test TUI export flow manually
3. Test error conditions (non-writable dir, empty db)
4. Test idempotent behavior (re-run export)
5. Test cross-platform paths

### Deliverables
- Test results documented
- Any bug fixes

### Definition of Done
- [ ] Manual CLI test passes
- [ ] Manual TUI test passes
- [ ] Edge cases handled
- [ ] Code review complete

## Phase 5: Documentation
**Goal:** Update all documentation

### Tasks
1. Update README.md with export feature
2. Update --help output
3. Update TUI help view with 'e' keybinding
4. Finalize PRD completion summary

### Deliverables
- Updated README.md
- Updated help views
- Completion summary

### Definition of Done
- [ ] README mentions export feature
- [ ] --help shows --export flag
- [ ] TUI help shows 'e' for export
- [ ] Completion summary written

## Dependencies Between Phases

```
Phase 1 (Foundation)
  ↓
Phase 2 (CLI) ──────┐
  ↓                  │
Phase 3 (TUI)        │
  ↓                  │
Phase 4 (Testing)    │
  ↓                  │
Phase 5 (Docs) ◄─────┘
```

## Timeline Estimate

| Phase | Estimated Time | Risk Level |
|-------|---------------|------------|
| 1 | 2 hours | Low |
| 2 | 1 hour | Low |
| 3 | 3 hours | Medium |
| 4 | 2 hours | Low |
| 5 | 1 hour | Low |
| **Total** | **9 hours** | **Low-Medium** |

## Resources Required

- Go 1.21+
- Existing dreams database for testing
- Write permissions for test directories

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Cross-platform path issues | Use filepath.Join exclusively |
| Large database performance | Test with 1000+ dream records |
| TUI state complexity | Keep export as simple modal flow |
| Partial write failures | Atomic file writes (temp + rename) |
