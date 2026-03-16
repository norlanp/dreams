# Dream Statistics Acceptance Checklist

Status: in-progress (todo #9)

This checklist maps dream-statistics requirements and acceptance criteria to automated coverage and targeted manual validation.

## Automated Coverage Matrix

| REQ | AC | Coverage Type | Test / Evidence |
| --- | --- | --- | --- |
| REQ-001 | AC-001 | Integration (TUI flow) | `internal/tui/analysis_e2e_test.go` `TestModelUpdate_ShouldValidateStatisticsNavigationCacheAndRerunFlow` (list -> `s` enters Dream Statistics) |
| REQ-002 | AC-002 | Integration (TUI flow) | `internal/tui/analysis_e2e_test.go` `TestModelUpdate_ShouldValidateStatisticsNavigationCacheAndRerunFlow` (cached analysis shown on first render) |
| REQ-003 | AC-003 | Integration (TUI flow) | `internal/tui/analysis_e2e_test.go` `TestModelUpdate_ShouldValidateStatisticsNavigationCacheAndRerunFlow` (press `r`, loading state, refresh with new cached result) |
| REQ-004 | AC-004 | Behavioral | `internal/tui/analysis_view_test.go` `TestModelUpdate_ShouldPersistAndRefreshAnalysisAfterSuccessfulRerun` (clusters + top terms + dream IDs parsed and persisted) |
| REQ-005 | AC-005 | Behavioral | `internal/tui/analysis_view_test.go` `TestModelUpdate_ShouldGuardMinimumDreamThresholdBeforeRunningPython` (guardrail blocks execution below threshold) |
| REQ-006 | AC-006 | Integration + unit | `internal/tui/analysis_e2e_test.go` `TestModelUpdate_ShouldShowLocalTimezoneTimestampWhenOpeningStatistics`; `internal/tui/analysis_view_test.go` `TestFormatAnalysisTimestamp_ShouldConvertUTCToLocalTime` |
| REQ-007 | AC-007 | Behavioral | `internal/tui/analysis_view_test.go` `TestModelUpdate_ShouldExposeRunnerFailureInAnalysisState` (execution failure surfaced, previous cache preserved) |

## Manual Acceptance Checks

Run these checks in a local environment with an existing `./dreams.db`.

| AC | Manual Check | Expected Result |
| --- | --- | --- |
| AC-001 | From list view, press `s` | Dream Statistics view appears immediately |
| AC-002 | Ensure cached analysis exists, open stats view | Last analyzed timestamp, dream count, and clusters are visible on initial render |
| AC-003 | In stats view, press `r` | `Running analysis...` appears, then view refreshes with latest persisted output |
| AC-005 | Reduce dataset below minimum (default 5), press `r` | No run executes; user sees too-few-dreams guidance |
| AC-006 | Compare displayed timestamp against known UTC analysis time | Displayed value is local-time converted and timezone-labeled |
| AC-007 | Force Python runner failure (non-zero exit), press `r` | Clear execution-failure message shown; prior cached analysis remains visible |

## Validation Command

```bash
go test ./...
```
