# Night Priming Execution Plan

## Objective
Deliver a keyboard-first Night Priming flow in the existing TUI that reliably produces pre-sleep priming content using strict fallback order (Personalized -> Community -> AI Generated -> Template), with resilient degraded behavior, local cache/log persistence, and targeted automated coverage for fallback, config, cache TTL, and error handling.

## Implementation Slices / Phases

### Phase 1 - Contracts and persistence foundation
- Define priming domain contracts (source label, outcome, attempt metadata) and fallback sequencing boundaries.
- Add storage migrations and repository methods for `priming_cache` (24h TTL) and `priming_logs`.
- Establish structured observability event shape for source attempt, fallback transition, cache hit/miss, and terminal failure.
- AC focus: AC-006, AC-009, AC-010 (schema/event preconditions).

### Phase 2 - Content sources and adapters
- Implement personalized source adapter backed by existing clustering output; skip cleanly when unavailable.
- Implement Reddit on-demand adapter for `r/LucidDreaming` with text-rich filtering and defensive parsing.
- Implement AI adapter with required env validation (`AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`) and optional one-time retry via `AI_MODEL_FALLBACK`.
- Implement static template source as deterministic terminal fallback.
- AC focus: AC-004, AC-005, AC-007, AC-008.

### Phase 3 - Orchestration and resilience behavior
- Implement generator orchestration with strict source order and explicit transition logic.
- Wire cache-first read path for Reddit source and network bypass on fresh entries.
- Record per-attempt observability and per-display priming log outcomes.
- Ensure non-blocking failures preserve retry/next cycle behavior.
- AC focus: AC-003, AC-006, AC-009, AC-010, AC-011.

### Phase 4 - TUI integration
- Add `nightView` entry from list view (`p`), exit (`esc`), and next-content (`n`) navigation.
- Render source label on every shown item and concise degraded-mode status when failures occur.
- Preserve list selection state when returning from night priming.
- AC focus: AC-001, AC-002, AC-011.

### Phase 5 - Verification and release hardening
- Add table-driven unit tests for fallback matrix and source-specific behavior.
- Add integration tests for TTL cache behavior, env validation failures, AI fallback model retry, and degraded-mode continuity.
- Run full suite and perform keyboard-driven acceptance walkthrough.
- AC focus: AC-012 and full AC regression.

## Dependency Graph Narrative
The critical chain is: persistence/contracts -> sources/adapters -> orchestrator -> TUI wiring -> integration verification.

Phase 1 must land first because cache/log tables and event contracts are consumed by orchestrator and tests. In Phase 2, source adapters can be developed in parallel except the AI adapter, which depends on shared config validation conventions. Phase 3 depends on all source interfaces from Phase 2 and persistence primitives from Phase 1. Phase 4 depends on generator behavior from Phase 3 to keep key handling simple and avoid TUI-specific fallback logic duplication. Phase 5 depends on all prior slices and should lock behavior against AC scenarios before rollout.

## Risk Handling Plan
- Reddit schema/rate-limit instability: use defensive parsing, filter only text-rich posts, prefer fresh cache before network, and fallback immediately on fetch/parse failure.
- AI misconfiguration/outage: fail fast on required env vars, retry once with `AI_MODEL_FALLBACK` when configured, then continue to template.
- Silent degradation reducing user trust: persist source/outcome records and emit structured attempt + transition events for diagnosis.
- TUI instability under repeated failures: keep failure handling non-blocking, surface concise status text, and keep `n` retry path always available.
- Test brittleness around external APIs: isolate adapters, use deterministic fixtures/stubs, and keep table-driven fallback coverage source-order focused.

## Test Strategy by Slice
- Phase 1: repository/migration tests for `priming_cache` TTL query behavior and `priming_logs` write/read invariants.
- Phase 2: unit tests per source adapter (personalized available/unavailable, Reddit filter rules, AI config validation + fallback model retry, template always available).
- Phase 3: table-driven orchestrator tests for full fallback matrix and observability/log side effects.
- Phase 4: TUI update-model tests for `p`, `n`, `esc`, source label rendering state, and return-to-list selection continuity.
- Phase 5: integration tests for cache-hit network bypass, invalid AI config degrade-to-template path, and all-dynamic-sources-fail responsiveness.

## Rollout / Verification Checklist
- [ ] Migrations apply cleanly on empty and populated local DB.
- [ ] `p` opens night priming, `esc` returns to list with prior selection preserved.
- [ ] `n` cycles content and each item shows correct source label.
- [ ] Fallback order verified as Personalized -> Community -> AI Generated -> Template.
- [ ] Fresh Reddit cache prevents outbound fetch and records cache hit.
- [ ] Missing AI config yields actionable message and continues to template.
- [ ] Source attempts, transitions, and final outcomes are persisted/logged.
- [ ] Automated tests covering fallback/cache/config/error paths pass in CI.
