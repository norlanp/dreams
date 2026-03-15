Meta: ID=PRD-dream-statistics | Status=approved | Size=M

Problem: Users of the Dreams CLI need fast, objective identification of recurring dream patterns (dreamsigns) from journal entries | Lucid dreamers and frequent journal users | Manual review is slow and inconsistent, reducing follow-through and insight quality
Scope: IN [x] Batch dream clustering from existing SQLite entries, cached results display in TUI, explicit re-run flow, local-time timestamping, minimum-data guardrails, and user-facing error states | OUT [ ] Real-time or streaming analysis, historical trend comparison across multiple runs, cloud sync, model training UI, and non-local storage backends

Requirements:
| REQ-001 | User can open Dream Statistics view from the dream list via keyboard-first navigation ('s') | P0 | GIVEN the user is in list view WHEN they press 's' THEN the app opens Dream Statistics view |
| REQ-002 | System displays cached latest analysis immediately when cache exists | P0 | GIVEN cached analysis exists WHEN Dream Statistics view loads THEN latest cached results and timestamp are shown without re-running ML |
| REQ-003 | User can explicitly re-run analysis from Dream Statistics view ('r') with loading state | P0 | GIVEN user is in Dream Statistics view WHEN they press 'r' THEN app shows loading and starts analysis pipeline asynchronously |
| REQ-004 | Analysis pipeline clusters dream text and returns top dreamsign terms per cluster | P0 | GIVEN sufficient dreams exist WHEN analysis runs THEN system outputs cluster count, top terms, and mapped dream IDs |
| REQ-005 | System enforces minimum dream count threshold before analysis | P0 | GIVEN stored dreams are below threshold WHEN user triggers analysis THEN app shows actionable too-few-dreams message and does not run clustering |
| REQ-006 | Analysis results are persisted for reuse and include UTC timestamp converted to local time for display | P0 | GIVEN analysis succeeds WHEN results are saved and later viewed THEN data is read from cache and timestamp is presented in local time |
| REQ-007 | Failures in DB access, Python execution, or JSON parse are surfaced with clear user feedback | P1 | GIVEN a failure occurs WHEN analysis or load is attempted THEN user sees a specific error state and app remains usable |

Acceptance Criteria:
| AC-001 | REQ-001 | GIVEN at least one dream exists WHEN user presses 's' in list view THEN Dream Statistics screen is shown |
| AC-002 | REQ-002 | GIVEN prior analysis exists WHEN Dream Statistics opens THEN cached timestamp, dreams analyzed count, and clusters are visible within the initial render |
| AC-003 | REQ-003 | GIVEN Dream Statistics is open WHEN user presses 'r' THEN loading indicator appears, run executes, and screen refreshes with latest results on success |
| AC-004 | REQ-004 | GIVEN 20+ dreams with recurring terms WHEN analysis completes THEN response includes one or more clusters with top terms and associated dream IDs |
| AC-005 | REQ-005 | GIVEN fewer than minimum dreams (default 5) WHEN user triggers analysis THEN no ML execution occurs and guidance message is shown |
| AC-006 | REQ-006 | GIVEN an analysis completed at UTC time T WHEN displayed in UI THEN timestamp is shown in user local timezone with correct conversion |
| AC-007 | REQ-007 | GIVEN Python process exits non-zero WHEN run is triggered THEN user sees execution failure message and prior cached results (if any) remain available |

APIs:
| API-001 | keypress | tui://list-view/s | REQ-001 |
| API-002 | query | sqlite://dream_analysis/latest | REQ-002, REQ-006 |
| API-003 | keypress | tui://analysis-view/r | REQ-003 |
| API-004 | command | cmd://python/extract_dreamsigns | REQ-004, REQ-005, REQ-007 |
| API-005 | write | sqlite://dream_analysis + sqlite://dream_clusters | REQ-006 |
