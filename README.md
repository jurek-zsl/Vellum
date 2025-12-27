# Vellum

Vellum is a CLI tool written in Go for managing and executing scripts and shell aliases with an elegant TUI interface.

## Features

- **Organize Scripts**: Manage scripts in subfolders.
- **Aliases**: Create and manage shell aliases.
- **TUI**: built with Bubble Tea for a smooth experience.
- **Logging**: Automatic logging of execution output with 7-day retention.
- **Form Wizards**: Interactive forms for creating, importing, and editing.
- **Configurable**: Change storage directory easily.

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from Source

```bash
git clone https://github.com/jurekzsl/vellum.git
cd vellum
go build -o vellum ./cmd/vellum
mv vellum /usr/local/bin/ # Optional
```

## Usage

Run the tool:

```bash
vellum
```

### Keybindings

- **Navigation**: Arrow keys (`↑`, `↓`)
- **Filter/Search**: Type to filter
- **`enter`**: Run selected script/alias
- **`a`**: Create new script
- **`i`**: Import existing script
- **`l`**: Add new alias
- **`e`**: Edit selected script/alias
- **`d`**: Delete selected script/alias
- **`c`**: Change Vellum storage directory
- **`h`**: Toggle help menu
- **`q`**: Quit

## Directory Structure

By default, Vellum stores data in `~/vellum/`:

- `scripts/`: Contains script files and metadata.
- `aliases/`: Contains alias metadata.

Logs are stored within each script/alias folder.

## Configuration

Configuration is stored in `~/.config/vellum/config.json`.

## Supported Types

- **Coding**: Python (.py), JS (.js), Ruby (.rb), Perl (.pl), PHP (.php), Lua (.lua), Tcl (.tcl), TS (.ts), JSX (.jsx), TSX (.tsx)
- **Shell**: Bash (.sh), Zsh (.zsh), Fish (.fish), PowerShell (.ps1)
- **Compiled**: Go (.go), Rust (.rs), Java (.java), Kotlin (.kt), Swift (.swift)
- **System**: Bat (.bat), Cmd (.cmd), VBS (.vbs), AppleScript (.applescript)

## License

MIT
