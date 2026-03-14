# AGENTS.md - Development Guidelines for Dreams

## Project Overview

CLI-based dream journal application with TUI interface for recording and analyzing dreams.

## Tech Stack

- **Language**: Go
- **TUI Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Styling**: Lipgloss (github.com/charmbracelet/lipgloss)
- **Database**: SQLite with modernc.org/sqlite
- **CLI**: Cobra or standard flag parsing

## Commands

### Development
```bash
go run main.go          # Run application
```

### Building
```bash
go build -o dreams      # Build binary
```

### Linting & Type Checking
```bash
go vet ./...            # Go vet
golangci-lint run       # Run linter (if installed)
gofmt -s -w .           # Format code
```

### Testing
```bash
go test ./...           # Run all tests
go test -v -run TestName  # Run single test
go test -v -run TestName -count=1  # Run single test without cache
```

## Code Style

### General Principles
- Minimal scope - implement only what's requested
- Fail fast - surface errors early
- No secrets - use env vars, never commit credentials
- Short functions - max ~60 lines, single responsibility
- No dynamic allocation after init

### Go Conventions
- Use `go.mod` for dependencies
- Follow standard Go project layout
- Error handling: `if err != nil { return err }`
- Naming: camelCase, PascalCase for acronyms (HTTP, URL, API)
- Interfaces defined by consumers
- Zero value literals preferred
- Check all errors, use all returns

### TUI Conventions (Bubble Tea)
- Handle terminal resize via tea.WindowSizeMsg
- Support keyboard navigation as primary input
- Use Bubble Tea's model pattern (Update, View)
- Show helpful prompts and confirmations
- Display loading states for async operations
- Use lipgloss for consistent styling

### Naming Conventions
- **Files**: snake_case (`dream_entry.go`)
- **Functions**: camelCase (`getDreams`, `createEntry`)
- **Types/Interfaces**: PascalCase (`DreamEntry`, `Model`)
- **Constants**: SCREAMING_SNAKE_CASE
- **Packages**: short, lowercase, no underscores (`tui`, `storage`, `model`)

### Imports
- Group: standard lib → external → internal
- Use explicit imports, avoid wildcard
```go
import (
    "context"
    "fmt"
    "time"

    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    "dreams/internal/model"
    "dreams/internal/storage"
)
```

### Formatting
- Run `gofmt -s` before committing
- Max line length: 100 characters
- Use `go fmt` for standard formatting

### Error Handling
- Never silently ignore errors
- Wrap errors with context: `fmt.Errorf("failed to create dream: %w", err)`
- Return typed errors for recoverable vs fatal
- Log errors at appropriate level
- Use sentinel errors for known failure cases

### Testing
- Use TDD: test first, then implement
- Test behavior, not implementation
- Use table-driven tests where applicable
- Test TUI components with model/state tests
- Follow naming: `TestFunctionName_ShouldDoSomething`
- Group tests with `t.Run`

## File Organization

```
cmd/
  main.go              # Entry point
internal/
  model/               # Data models
    dream.go
  tui/                 # TUI components
    model.go           # Main TUI model
    views.go           # View functions
    keys.go            # Key bindings
  storage/             # Database/file handling
    sqlite.go
    repository.go
pkg/                   # Public libraries (optional)
docs/                  # Documentation
tests/                 # Test files (if external)
```

## Git Conventions
- Commit format: `type: description` (feat/fix/docs/refactor/test/chore/style/perf)
- Keep commits small and focused
- Never commit secrets or env files

## Testing Guidelines
- Test the model/state, not the view
- Test keyboard input handling via tea.Model
- Test state transitions
- Group tests with `t.Run`
- Follow naming: `TestFunctionName_ShouldDoSomething`
- Use golden files for view comparison if needed

## Data Storage
- Use SQLite for local persistence
- Use modernc.org/sqlite driver (pure Go)
- Store in appropriate app data directory
- Handle migration for schema changes
- Never commit database files
- Use `./var/` directory for data (add to .gitignore)

## Environment Variables
- Use `.env` with `.env.example` for template
- Never commit `.env` files
- Validate required vars at startup

## Code Review Checklist

- [ ] Code compiles without errors (`go build`)
- [ ] No vet warnings (`go vet ./...`)
- [ ] Tests pass (`go test ./...`)
- [ ] Formatted (`gofmt -s`)
- [ ] Error handling in place
- [ ] No secrets in code