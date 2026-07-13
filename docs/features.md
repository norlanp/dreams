# Features

Feature status is one of `planned`, `active`, `deprecated`, `removed`, or `unscoped`. Active features are currently supported.

| Feature | Status | Scope | Size |
| --- | --- | --- | --- |
| create-dream | unscoped | - | - |
| list-dreams | unscoped | - | - |
| view-dream | unscoped | - | - |
| delete-dream | unscoped | - | - |
| search-dreams | unscoped | - | - |
| dream-statistics | active | [Details](#dream-statistics) | M |
| export-markdown | active | [Details](#export-markdown) | M |
| night-priming | active | [Details](#night-priming) | L |

## Dream Statistics

Analyze recurring dream patterns from the list view.

- Open with `s`; re-run analysis with `r`.
- Show the latest cached analysis immediately, including local-time timestamp, dream count, clusters, top terms, and associated dream IDs.
- Run Python clustering asynchronously on demand; retain prior cached results when a new run fails.
- Require the configured minimum dream count before running analysis.
- Surface database, Python execution, and JSON parsing failures without leaving the TUI unusable.
- Store completed analysis and cluster data in SQLite for later display.

Validation covers navigation, cached loading, re-run behavior, minimum-data handling, local-time conversion, persistence, and runner failures.

## Export Markdown

Export all stored dreams as individual Markdown files.

- Run `dreams --export <directory>` from the CLI or press `e` in the list view.
- Create missing output directories and write each file atomically.
- Name files as `YYYY-MM-DD-HH-MM-SS-<dream-id>-dream.md` and include YAML frontmatter with the dream date.
- Preserve dream content, including multiline content and special characters.
- Return the exported count and report failures with context.
- Accept only export directories under the current working directory.

Validation covers file naming and content, directory creation, atomic writes, empty data, write failures, path validation, and CLI and TUI flows.

## Night Priming

Provide a keyboard-first pre-sleep priming flow.

- Open with `p`, request the next item with `n`, and return to the prior list selection with `esc`.
- Use strict fallback order: Personalized, Community, AI Generated, then Template.
- Generate personalized content from saved clustering output when available.
- Fetch text-focused r/LucidDreaming content on demand and reuse fresh local cache entries for 24 hours.
- Use an OpenAI-compatible provider when `AI_BASE_URL`, `AI_API_KEY`, and `AI_MODEL` are configured; retry once with `AI_MODEL_FALLBACK` when set.
- Record source attempts, fallback transitions, cache outcomes, and displayed-content outcomes.
- Keep the view responsive when a source fails and present concise degraded-mode feedback.

Validation covers source ordering, cache TTL and network bypass, AI configuration and fallback, persistence, and non-blocking TUI error behavior.
