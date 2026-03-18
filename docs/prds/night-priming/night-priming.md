Meta: ID=PRD-night-priming | Status=completed | Size=L

## Problem
Users want a short, reliable pre-sleep priming flow in the TUI that increases dream recall and lucidity cues the next morning. Today there is no dedicated night mode, no source cascade for priming content, and no resilient handling for network/API failures, which makes priming inconsistent and hard to trust.

## Scope
IN:
- Add a keyboard-first night priming view in the existing TUI.
- Deliver content via strict fallback order: personalized -> Reddit on-demand -> AI-generated -> static template.
- Support OpenAI-compatible provider configuration via `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`, `AI_MODEL_FALLBACK`.
- Add local caching for Reddit content, source usage logging, and basic observability.
- Add automated tests for fallback behavior, configuration, cache, and error handling.

OUT:
- Background sync jobs, schedulers, or daemons.
- New cloud services, user accounts, or remote persistence.
- Ranking/ML optimization of source effectiveness.
- Multi-provider orchestration beyond one OpenAI-compatible endpoint.

## Requirements
| ID | Requirement | Priority | Scenario (GIVEN/WHEN/THEN) |
| --- | --- | --- | --- |
| REQ-001 | User can open and exit Night Priming mode from list view using keyboard-first controls (`p` to open, `esc` to close). | P0 | GIVEN the user is in list view WHEN they press `p` THEN night priming view opens fullscreen, and GIVEN night priming is open WHEN they press `esc` THEN the app returns to list view. |
| REQ-002 | Night Priming view supports next-content navigation (`n`) and always shows source label (`Personalized`, `Community`, `AI Generated`, `Template`). | P0 | GIVEN night priming is open WHEN user presses `n` THEN next priming item is shown with its source label. |
| REQ-003 | Generator uses strict fallback order: personalized dream-sign content first, then Reddit, then AI generation, then static templates. | P0 | GIVEN a request for priming content WHEN higher-priority source is unavailable THEN system attempts the next source in order until content is produced. |
| REQ-004 | Personalized source uses existing clustering output when available and skips without error when unavailable. | P0 | GIVEN clusters exist WHEN priming content is requested THEN personalized content is generated from top dream signs, and GIVEN clusters do not exist THEN source is skipped and fallback continues. |
| REQ-005 | Reddit content is fetched on-demand from r/LucidDreaming and filtered to text-rich posts suitable for priming. | P0 | GIVEN Reddit is reachable WHEN priming requests community content THEN app fetches posts on demand, filters low-signal entries, and returns readable priming text. |
| REQ-006 | Reddit results are cached locally with TTL (24h) and reused before network fetch when cache is fresh. | P0 | GIVEN cached Reddit entries are younger than 24h WHEN priming needs community content THEN cached entries are used instead of making a network request. |
| REQ-007 | AI generation uses OpenAI-compatible chat completions with required env config and optional fallback model. | P0 | GIVEN `AI_BASE_URL`, `AI_API_KEY`, and `AI_MODEL` are valid WHEN AI generation is needed THEN app calls chat completions, and WHEN primary model fails and `AI_MODEL_FALLBACK` is set THEN app retries once with fallback model. |
| REQ-008 | Missing or invalid AI configuration fails fast with clear, actionable error and continues fallback to lower-priority source when possible. | P0 | GIVEN AI config is missing/invalid WHEN AI source is attempted THEN user-facing error state is recorded and system proceeds to static template source. |
| REQ-009 | Displayed priming events are logged locally with timestamp, selected source, and outcome (success/failure). | P1 | GIVEN a priming item is shown WHEN render succeeds or fails THEN an entry is persisted for later analysis/auditing. |
| REQ-010 | System emits basic observability signals for source attempts, fallback transitions, cache hit/miss, and terminal failure. | P1 | GIVEN priming generation executes WHEN each source attempt occurs THEN structured logs/events capture attempt order and result. |
| REQ-011 | Error handling is non-blocking: if one source fails, app remains usable, shows concise status, and allows retry/next. | P0 | GIVEN a source error occurs WHEN user is in night priming view THEN app does not crash, communicates status, and user can press `n` to retry next content cycle. |
| REQ-012 | Test coverage includes table-driven unit tests for source selection and integration tests for cache/config/error paths. | P0 | GIVEN automated test suite runs WHEN night priming module tests execute THEN fallback order, config validation, cache TTL, and failure handling are verified. |

## Acceptance Criteria
| ID | REQ Ref | Scenario (GIVEN/WHEN/THEN) |
| --- | --- | --- |
| AC-001 | REQ-001 | GIVEN list view is active WHEN user presses `p` THEN night priming view is shown, and WHEN user presses `esc` THEN list view is restored with prior selection intact. |
| AC-002 | REQ-002 | GIVEN night priming view is open WHEN user presses `n` twice THEN two new items are shown and each render includes a source label. |
| AC-003 | REQ-003 | GIVEN personalized and Reddit both fail and AI succeeds WHEN content is requested THEN final displayed item is from AI and attempt order follows configured fallback sequence. |
| AC-004 | REQ-004 | GIVEN dream clusters exist WHEN personalized source runs THEN generated text includes at least one extracted dream-sign term. |
| AC-005 | REQ-005 | GIVEN Reddit returns mixed post types WHEN community source processes response THEN only text-appropriate entries are eligible for display. |
| AC-006 | REQ-006 | GIVEN fresh Reddit cache exists WHEN community source is requested THEN no outbound Reddit request is made and cache hit is logged. |
| AC-007 | REQ-007 | GIVEN valid AI env vars and failing primary model with fallback configured WHEN AI generation is attempted THEN second attempt uses `AI_MODEL_FALLBACK` and returns content on success. |
| AC-008 | REQ-008 | GIVEN `AI_API_KEY` is missing WHEN AI source is attempted THEN user sees actionable configuration error and pipeline continues to template fallback. |
| AC-009 | REQ-009 | GIVEN any source produces output WHEN item is shown THEN one priming log record is persisted with timestamp, source, and success state. |
| AC-010 | REQ-010 | GIVEN fallback occurs from Reddit to AI WHEN generation completes THEN observability output contains both attempt results and fallback transition. |
| AC-011 | REQ-011 | GIVEN all dynamic sources fail WHEN user presses `n` THEN app remains responsive, shows template content or terminal error message, and allows another retry. |
| AC-012 | REQ-012 | GIVEN CI test run WHEN night priming tests execute THEN designated unit and integration suites pass and cover fallback, cache TTL, env validation, and degraded-mode behavior. |

## APIs
| ID | Method | Path | REQ Ref |
| --- | --- | --- | --- |
| API-001 | keypress | `tui://list-view/p` | REQ-001 |
| API-002 | keypress | `tui://night-view/n`, `tui://night-view/esc` | REQ-002, REQ-011 |
| API-003 | query | `sqlite://dream_analysis/latest` | REQ-004 |
| API-004 | GET | `https://www.reddit.com/r/LucidDreaming/.json` | REQ-005 |
| API-005 | query/write | `sqlite://priming_cache` | REQ-006 |
| API-006 | POST | `${AI_BASE_URL}/chat/completions` | REQ-007, REQ-008 |
| API-007 | write | `sqlite://priming_logs` | REQ-009 |
| API-008 | event | `obs://priming/source-attempt` | REQ-010 |

## Non-Goals
- No autoplay timer configuration or session scheduling in this iteration.
- No personalized effectiveness scoring, A/B testing, or recommendation tuning.
- No support for additional community sources beyond Reddit.
- No changes to dream editing, export, or statistics UX beyond integration touchpoints required for priming.

## Risks & Mitigations
- Reddit API rate limiting or schema drift -> Use cache-first behavior, defensive parsing, and automatic fallback to AI/template.
- AI provider outages or invalid credentials -> Fail fast on config, retry once with fallback model, then degrade gracefully to template.
- Low-quality or inappropriate fetched content -> Apply strict filtering rules and sanitize displayed text.
- Silent failures reduce trust -> Persist source/outcome logs and emit structured observability events for debugging.
- Test brittleness around external calls -> Use deterministic fixtures/mocks and table-driven tests for fallback matrix.
