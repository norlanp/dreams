# dreams

A CLI dream journal with a beautiful TUI interface. Record, view, and analyze your dreams with an intuitive keyboard-driven interface.

## Features

- **TUI Interface**: Built with Bubble Tea for a smooth terminal experience
- **SQLite Storage**: Dreams persist locally in a SQLite database
- **Keyboard Navigation**: Full keyboard control for quick entry and navigation
- **View & Edit**: Browse your dream history and edit entries

## Tech Stack

- **Go** 1.25+
- **Bubble Tea** - TUI framework
- **Lipgloss** - Styling
- **SQLite** - Local database (via modernc.org/sqlite)

## Quick Start

```bash
# Run in development mode
go run cmd/main.go

# Or use Make
make dev
```

## Build

```bash
# Build binary
make build

# Run the built binary
./build/dreams
```

## Development

```bash
# Run tests
make test

# Lint and format
make lint
make fmt

# Clean build artifacts
make clean
```

## Project Structure

```
cmd/main.go              # Entry point
internal/
  model/                 # Data models
    dream.go
  tui/                   # TUI components
    model.go
    update.go
    views.go
  storage/               # Database layer
    repository.go
    sqlc/               # Generated SQLC code
var/                     # Database storage (gitignored)
```

## Usage

Launch the application to enter the TUI. Use keyboard shortcuts to navigate:

- `tab` - Switch between views
- `enter` - Select/confirm
- `esc` - Go back/cancel
- `ctrl+c` - Quit

## License

MIT
