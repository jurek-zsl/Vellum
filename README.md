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

**brew and other would be added after exiting beta**

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
- **Filter/Search**: Focus on filter bar (`/`)
- **`enter`**: Run selected script/alias
- **`r`**: Run scipt/alias with advanced options
- **`a`**: Create new script
- **`i`**: Import existing script
- **`l`**: Add new alias
- **`e`**: Edit selected script/alias
- **`d`**: Delete selected script/alias
- **`c`**: Copy Script or Alias to clipboard
- **`o`**: Open Script or Alias subfolder in Explorer
- **`s`**: Cycle through sorting modes
- **`q`**: Quit

## Directory Structure

By default, Vellum stores data in `~/vellum/`:

- `scripts/`: Contains script files and metadata.
- `aliases/`: Contains alias metadata.

Logs are stored within each script/alias folder.

## Configuration

Configuration is stored in `~/.config/vellum/config.json`.

### Configuration Parameters

> [!IMPORTANT]
> If you're using macOS setup `file_manager` to `open`. If you're using Linux setup `file_manager` to `xdg-open`. If you're using Windows setup `file_manager` to `explorer`. Without that `o` key will not work.

| Parameter | Type | Default | Description | Options / Examples |
| :--- | :--- | :--- | :--- | :--- |
| `theme` | String | `"#BD93F9"` | The hex color code used for the TUI theme. | `"#00FFFF"` (Cyan), `"#FF0000"` (Red) |
| `default_shell` | String | `$SHELL` | The shell used to execute commands and scripts. Important for process substitution support. | `"/bin/zsh"`, `"/bin/bash"`, `"/usr/bin/fish"` |
| `default_editor` | String | `$EDITOR` | The command to open the external text editor for editing scripts. | `"nano"`, `"vim"`, `"code"`, `"micro"` |
| `sort_order` | String | `"name"` | Controls how items are sorted in the list. | `"name"` (A-Z), `"lastrun"` (Most recent first), `"type"` (Group by type) |
| `log_retention_days`| Integer | `7` | Number of days to keep log files before auto-deletion. | `7`, `30`, `365` |
| `max_log_files` | Integer | `50` | Maximum number of log files to keep. Oldest logs are deleted when this limit is exceeded. | `50`, `100`, `1000` |
| `file_manager` | String | Auto | The command to open directories (used by `o` key). | `"open"` (macOS), `"xdg-open"` (Linux), `"explorer"` (Windows) |
| `lazy_load` | Boolean | `false` | Performance optimization. If true, defers loading heavy resources. | `true`, `false` |
| `cache_enabled` | Boolean | `true` | Enables caching of script metadata for faster startup. | `true`, `false` |

## Supported Types

> [!NOTE]
> For now there are only few script types available. I'm going to add more script types with each update.

- **Coding**: `Python (.py)`, `JavaScript(.js)`, ~`Ruby (.rb)`~, ~`Perl (.pl)`~, ~`PHP (.php)`~, ~`Lua (.lua)`~, ~`Tcl (.tcl)`~, ~`TS (.ts)`~, ~`JSX (.jsx)`~, ~`TSX (.tsx)`~
- **Shell**: `Bash (.sh)`, ~`Fish (.fish)`~, ~`PowerShell (.ps1)`~
- **Compiled**: `Go (.go)`, ~`Rust (.rs)`~, ~`Java (.java)`~, ~`Kotlin (.kt)`~, ~`Swift (.swift)`~
- **System**: ~`Bat (.bat)`~, ~`Cmd (.cmd)`~, ~`VBS (.vbs)`~, ~`AppleScript (.applescript)`~

## License

MIT
