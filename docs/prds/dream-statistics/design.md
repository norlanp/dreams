# Dreamsign Extraction via ML Clustering

**Status:** Brainstormed  
**Date:** 2026-03-14  
**Participants:** User, AI

## Problem Statement

Users want to identify recurring patterns (dreamsigns) in their dreams to improve lucid dreaming awareness. Manual review is time-consuming and subjective. This feature automatically extracts common themes using machine learning clustering.

## Goals

- Automatically identify recurring dream themes (dreamsigns)
- Provide insights without manual review of all entries
- Support lucid dreaming practice by highlighting patterns
- Handle batch processing triggered from CLI
- Gracefully handle small datasets (minimum threshold)
- Cache results to avoid re-running ML pipeline unnecessarily

## Approach

**Selected:** Hybrid batch processing with Python ML + Go TUI integration

Go CLI triggers Python analysis, which reads from SQLite, performs clustering, and outputs JSON. Results are cached in SQLite and displayed in TUI with local time conversion.

### Alternatives Considered

- **Pure Go**: Go ML libraries exist but are less mature than scikit-learn. Would sacrifice accuracy and maintainability.
- **Real-time processing**: Too slow for ML; clustering should be batch-based.
- **Embedded Python**: Would complicate the build with cgo dependencies.

## Architecture

### Components

```
internal/analysis/           # Python ML code
├── pyproject.toml          # uv project config
├── scripts/
│   └── extract_dreamsigns.py  # Main clustering script

internal/tui/               # Go TUI (existing)
├── model.go
├── update.go
└── views.go

var/                        # Data storage (existing)
└── dreams.db              # SQLite database
```

### Data Flow

1. User accesses analysis view via menu option from list view (press 's')
2. Go checks for cached analysis results in SQLite
3. If cached results exist:
   - Display cached results with analysis timestamp (converted from UTC to local time)
   - Show option to re-run analysis
4. If no cached results or user chooses to re-run:
   - Show loading indicator
   - Execute Python script via `exec.Command`
   - Python reads dreams from SQLite and performs clustering
   - Go parses JSON and stores results in SQLite
   - Display results with timestamp

### Technical Details

**Python Side:**
- scikit-learn for TF-IDF + K-means
- Auto-determines optimal cluster count using silhouette score
- Minimum 5 dreams required (configurable)
- Extracts top 5 terms per cluster as dreamsigns
- Outputs JSON with cluster info and dream IDs

**Go Side:**
- New `analysisView` state
- Cached results stored in SQLite (`dream_analysis`, `dream_clusters` tables)
- Analysis timestamp stored as UTC, displayed in local time
- Async execution with loading indicator when re-running
- Results view showing:
  - Analysis timestamp (local time)
  - Number of dreams analyzed
  - Number of clusters found
  - Top dreamsigns per cluster
  - Dreams in each cluster
- Menu trigger: 's' key from list view

**Caching Strategy:**
- Always display cached results first if available
- Show datetime in local timezone (converted from UTC)
- Allow explicit re-run via keybinding ('r')
- Store results in SQLite for instant display

## Error Handling

- Database not found → Show setup instructions
- Too few dreams → Explain minimum requirement
- Python execution failure → Show error message
- JSON parse failure → Log to debug file
- No cached results → Show empty state with prompt to run analysis

## Testing Strategy

- Unit tests for Python clustering with sample data
- Integration test with actual database
- Mock Python output for Go testing
- Minimum dreams threshold test
- Test UTC to local time conversion
- Test cache retrieval and storage

## Clarifications

**Trigger Method:** Menu option from list view using 's' key

**Caching Behavior:**
- Always show cached results if available
- Display analysis timestamp in local time (stored as UTC in database)
- Allow user to see when the last Python pipeline was run

**Display:** Show latest analysis only (no historical comparison for now)

## Next Steps

1. Add TUI view for displaying cached analysis results
2. Add 's' keybinding from list view to trigger analysis view
3. Add async command to execute Python and store results
4. Implement local time conversion for timestamps
5. Add option to re-run analysis ('r' keybinding)
6. Add tests for caching and time conversion
7. Update documentation
