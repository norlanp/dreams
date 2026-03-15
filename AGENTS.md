# Dreams - Agent Guidelines

CLI dream journal with Bubble Tea TUI + SQLite.

## Commands

```bash
go run main.go              # Dev
go test ./...               # Tests
go vet ./... && gofmt -s -w .  # Lint & format
go build -o dreams          # Build
```

## Code Style

- Max 60 lines/functions, single responsibility
- TDD: failing test → code → refactor
- `if err != nil { return fmt.Errorf("context: %w", err) }`
- Never ignore errors, never commit secrets

## Naming

- Files: `snake_case.go`
- Functions: `camelCase`
- Types: `PascalCase`
- Constants: `SCREAMING_SNAKE_CASE`
- Packages: short, lowercase

## Imports

```go
import (
    "std/lib"

    "github.com/external/pkg"

    "dreams/internal/pkg"
)
```

## Structure

```
cmd/main.go
internal/
  model/        # Data types
  tui/          # Bubble Tea components
  storage/      # SQLite
var/            # Data dir (gitignored)
```

## Testing

- Test behavior, not implementation
- Table-driven tests
- `TestFunctionName_ShouldDoSomething`
- Test state/model, not view rendering

## TUI (Bubble Tea)

- Handle `tea.WindowSizeMsg`
- Keyboard-first navigation
- Model pattern: `Init`, `Update`, `View`
- Show loading states for async ops

## Data

- SQLite with `modernc.org/sqlite`
- Store in `./var/`
- Migrations for schema changes

## Git

- Format: `type: description` (feat/fix/docs/refactor/test/chore/style/perf)
- Confirm before committing
