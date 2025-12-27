# Vellum

> [!IMPORTANT]
> WIP - Project is in very early stages of building, be aware of bugs and issues.

Vellum is a multi platform CLI tool written in Go for managing and executing scripts and shell aliases with an nice TUI interface.

## Platform Statuses

- **Windows 10/11**: Untested
- **macOS**: Tested on *macOS 26.2* – works (beta)
- **Debian-based Linux (apt)**: Untested (Probably works)
- **Arch-based Linux (pacman)**: Untested
- **Red Hat-based Linux (dnf/rpm)**: Untested

## Features

- **Organize Scripts**: Manage scripts in subfolders.
- **Aliases**: Create and manage shell aliases.
- **TUI**: built with Bubble Tea.
- **Logging**: Automatic logging of execution output with 7-day retention.
- **Form Wizards**: Interactive forms for creating, importing, and editing.

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
- **`c`**: Copy Script or Alias to clipboard
- **`o`**: Open Script or Alias subfolder in Explorer
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

> [!NOTE]
> For now there are only few script types available. I'm going to add more script types with each update.

- **Coding**: `Python (.py)`, `JavaScript(.js)`, ~`Ruby (.rb)`~, ~`Perl (.pl)`~, ~`PHP (.php)`~, ~`Lua (.lua)`~, ~`Tcl (.tcl)`~, ~`TS (.ts)`~, ~`JSX (.jsx)`~, ~`TSX (.tsx)`~
- **Shell**: `Bash (.sh)`, ~`Zsh (.zsh)`~, ~`Fish (.fish)`~, ~`PowerShell (.ps1)`~
- **Compiled**: `Go (.go)`, ~`Rust (.rs)`~, ~`Java (.java)`~, ~`Kotlin (.kt)`~, ~`Swift (.swift)`~
- **System**: ~`Bat (.bat)`~, ~`Cmd (.cmd)`~, ~`VBS (.vbs)`~, ~`AppleScript (.applescript)`~

## License

MIT
