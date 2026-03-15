# Execution Plan - dream-statistics

## Objective
Deliver a cache-first Dream Statistics experience in the CLI that lets users open analysis from the dream list, immediately see the latest stored results, explicitly re-run clustering on demand, and receive clear feedback for minimum-data and runtime failures, while keeping timestamps correct in local time.

## Workstreams
- ML Pipeline (Python): Maintain and harden dreamsign extraction (TF-IDF + clustering), minimum-dream threshold checks, and machine-readable JSON output.
- Persistence (SQLite): Use dream_analysis and dream_clusters for latest-result caching, retrieval, and overwrite semantics on re-run.
- TUI Flow (Go/Bubble Tea): Add keyboard navigation (s to open, r to rerun), loading/empty/error/result states, and local-time timestamp display.
- Execution Bridge (Go to Python): Run Python asynchronously, handle non-zero exits and parse failures, and keep app usable when execution fails.
- Quality and Validation: Add behavioral tests for thresholds, cache behavior, time conversion, and error handling; verify all PRD acceptance criteria.

## Ordered Implementation Phases
1. Stabilize completed foundation
   - Confirm completed infrastructure remains aligned (Python env/script, schema, storage methods).
   - Lock JSON contract between Python output and Go parser.
2. Wire navigation and analysis screen entry
   - Add or verify list-view s keybinding and transition to analysis view.
   - Ensure analysis view can initialize with cached-or-empty model state.
3. Implement cache-first read path
   - Load latest cached analysis on view open.
   - Render timestamp in local timezone and show counts and clusters on first render.
4. Implement explicit re-run path
   - Add r key handling with async loading state.
   - Enforce minimum-dream threshold before Python execution.
   - Execute Python, parse JSON, persist latest results, then refresh view.
5. Complete resilient UI states
   - Add distinct messages for too few dreams, DB access errors, Python execution failures, and JSON parse failures.
   - Preserve prior cached results when re-run fails.
6. Validate and harden
   - Add or expand tests across pipeline, cache retrieval and storage, threshold behavior, and timezone conversion.
   - Run end-to-end checks against acceptance criteria and finalize docs updates.

## Risks and Mitigations
- Risk: Python runtime or dependency mismatch breaks execution.
  Mitigation: Keep uv-based environment checks in test flow; surface actionable execution errors in TUI.
- Risk: JSON contract drift between Python output and Go parser.
  Mitigation: Define stable output schema and add parser contract tests with fixture payloads.
- Risk: Incorrect local-time conversion causes misleading last-run display.
  Mitigation: Add deterministic tests with fixed UTC inputs and expected local outputs.
- Risk: Re-run failure wipes useful prior insights.
  Mitigation: Treat cache as last-known-good; only replace persisted results after successful parse and store.
- Risk: Small datasets trigger low-quality clustering or confusing UX.
  Mitigation: Enforce threshold guardrail pre-execution and show specific guidance text.

## Validation Plan
- Unit tests (Python): clustering output shape, minimum threshold behavior, deterministic fixture expectations.
- Unit tests (Go storage/model): latest-analysis read and write behavior, cluster mapping, UTC persistence and local conversion formatting.
- TUI behavior tests: s enters analysis view, r triggers loading then success or error transitions, cached results display on initial render.
- Integration tests: run pipeline against test SQLite DB and verify persisted tables and refreshed analysis view model.
- Acceptance verification: map REQ and AC 001-007 to test cases and a short manual checklist for keyboard navigation and error-state clarity.
