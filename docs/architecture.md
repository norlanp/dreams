# Dreams Architecture

## Overview

Dreams is a CLI dream journal application with a Terminal User Interface (TUI) built using Bubble Tea. It provides dream recording, analysis, and export capabilities with local SQLite storage.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      TUI Layer (Bubble Tea)                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │  List    │ │  Detail  │ │  Create  │ │  Analysis    │   │
│  │  View    │ │  View    │ │  View    │ │  View        │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Services                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Export     │  │   Analysis   │  │   Night Priming  │  │
│  │   Service    │  │   Pipeline   │  │   Generator      │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Storage Layer (Repository)               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Repository Interface + SQLite Implementation        │   │
│  │  - Dream CRUD                                        │   │
│  │  - Analysis Storage                                  │   │
│  │  - Priming Content Cache                             │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Layer (SQLite)                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │  dreams  │ │ analysis │ │ clusters │ │  priming_*   │   │
│  │  table   │ │  table   │ │  table   │ │  tables      │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### TUI Layer (`internal/tui/`)

**Responsibilities:**
- Render terminal interface using Bubble Tea
- Handle user input and navigation
- Manage application state (views, selections, forms)
- Coordinate async operations via Tea commands

**Key Components:**
- `model.go` - Application model and state
- `update.go` - Event handling and state transitions
- `views.go` - View rendering functions

**Design Patterns:**
- Elm architecture (Model-View-Update)
- State machine for view transitions
- Command pattern for async operations

### Storage Layer (`internal/storage/`)

**Responsibilities:**
- Database connection management
- Migration execution
- Repository pattern implementation
- Data mapping between SQL and domain models

**Key Components:**
- `repository.go` - Repository implementation
- `db_path.go` - Database path resolution
- `sqlc/` - Generated SQL queries
- `migrations/` - Database schema migrations

**Design Patterns:**
- Repository pattern
- Unit of Work (transactional operations)
- Migration-based schema evolution

### Analysis Pipeline (`internal/analysis/`)

**Responsibilities:**
- Dream pattern analysis
- Clustering using ML techniques
- Dream sign extraction

**Implementation:**
- Python-based ML pipeline
- Invoked via `uv run` from Go
- Results stored in database

**Integration:**
- Go spawns Python process
- JSON-based communication
- Async execution with timeout handling

### Export Service (`internal/export/`)

**Responsibilities:**
- Export dreams to Markdown files
- Atomic file writes
- Path validation

**Design:**
- Stateless service functions
- Temp file + rename pattern
- Path traversal protection

### Night Priming (`internal/priming/`)

**Responsibilities:**
- Generate bedtime priming content
- Multiple content sources with fallback
- Cache management

**Sources:**
1. AI-generated (primary)
2. Personalized content
3. Community content
4. Template fallback

**Design Patterns:**
- Chain of Responsibility for source fallback
- Strategy pattern for content sources
- Caching with TTL

## Data Model

### Core Entities

**Dream**
```
- ID: int64
- Content: string (max 100KB)
- CreatedAt: time.Time
- UpdatedAt: time.Time
```

**Analysis**
```
- ID: int64
- AnalysisDate: time.Time
- DreamCount: int64
- NClusters: int64
- ResultsJSON: string
- CreatedAt: time.Time
```

**Cluster**
```
- ID: int64
- AnalysisID: int64 (FK)
- ClusterID: int64
- DreamCount: int64
- TopTerms: []string (JSON)
- DreamIDs: []int64 (JSON)
- CreatedAt: time.Time
```

**PrimingContent**
```
- ID: int64
- Source: string
- Title: string
- Content: string
- Category: string
- URL: string
- CreatedAt: time.Time
- UpdatedAt: time.Time
```

## Security Considerations

### Database Path
- Symlink validation prevents attacks
- Unsafe location detection (/etc, /proc, /sys, /dev, /tmp)
- Environment variable override with validation

### Export
- Absolute path resolution
- No path traversal sequences allowed
- Atomic writes prevent partial files

### API Keys
- Loaded from environment variables
- Sanitized from error responses
- Never logged or displayed

## Configuration

### Environment Variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `DREAMS_DB_PATH` | Database location | No |
| `DREAMS_EDITOR` | External editor | No |
| `AI_BASE_URL` | AI API endpoint | For AI priming |
| `AI_API_KEY` | AI authentication | For AI priming |
| `AI_MODEL` | AI model name | For AI priming |
| `AI_MODEL_FALLBACK` | Fallback model | No |

### File Structure

```
dreams/
├── cmd/main.go              # Entry point
├── internal/
│   ├── model/               # Domain models
│   ├── tui/                 # TUI components
│   ├── storage/             # Database layer
│   │   ├── migrations/      # SQL migrations
│   │   ├── queries/         # SQLC queries
│   │   └── sqlc/            # Generated code
│   ├── analysis/            # Python ML pipeline
│   ├── priming/             # Content generation
│   └── export/              # Export service
├── var/                     # Runtime data
│   └── dreams.log           # Application logs
└── docs/                    # Documentation
```

## Error Handling

### Principles
- All errors wrapped with context
- No ignored errors
- User-facing messages separated from logs
- Timeouts on all async operations

### Error Flow
```
Storage Error → Repository wraps → Service adds context → TUI displays
```

## Testing Strategy

### Unit Tests
- Model JSON marshaling
- Repository operations
- Service functions

### Integration Tests
- End-to-end export
- Database operations
- TUI message handling

### Test Patterns
- Table-driven tests
- Behavior testing over implementation
- State verification

## Performance Considerations

### Database
- Single connection (SQLite limitation)
- Connection pooling configured
- Transactional batch operations

### TUI
- Lazy loading of dream lists
- Async analysis execution
- Loading states for UX

### Analysis Pipeline
- 30-second timeout
- Context cancellation support
- Graceful degradation

## Extension Points

### Adding New Views
1. Add state to `viewState` type
2. Implement handler in `update.go`
3. Implement renderer in `views.go`
4. Add navigation from existing views

### Adding New Storage Operations
1. Add query to `internal/storage/queries/`
2. Run `sqlc generate`
3. Implement in `repository.go`
4. Add to repository interface

### Adding Priming Sources
1. Implement `Source` interface
2. Add to generator source chain
3. Implement fallback logic

## Deployment

### Development
```bash
go run cmd/main.go
```

### Production Build
```bash
go build -o dreams cmd/main.go
```

### Database Migrations
- Applied automatically on startup
- Embedded in binary via `embed.FS`
- No manual migration needed

## Monitoring

### Logs
- Location: `./var/dreams.log`
- Rotation: 10MB max (auto-archive to `.old`)
- Format: Standard log with timestamps

### Debugging
- Set `DREAMS_DB_PATH` for custom DB location
- Check `dreams.log` for errors
- Use `--export` flag for CLI export
