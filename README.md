# Vellum

Vellum is a modern, cross-platform CLI tool and interactive TUI written in Go for organizing, managing, and executing scripts and shell aliases with automated logging, log retention, and integrated log viewing.

## Features

- **TUI & CLI Dual Mode**: Interactive Bubble Tea interface or headless CLI execution (`vellum run <name>`).
- **Organize Scripts**: Manage scripts (Bash, Python, Go, Node.js, Swift, Ruby, PHP, and more) with automatic metadata tracking.
- **Shell Aliases**: Manage and export shell aliases safely (supporting Bash, Zsh, and Fish).
- **Integrated Log Viewer**: View historical run outputs directly inside the TUI with exit code badges and execution durations.
- **7-Day Automated Log Cleanup**: Automatic, resource-efficient rotation and retention of old log files.
- **Form Wizards**: Interactive forms built with Huh for creating, importing, editing, and running scripts with advanced parameters.
- **Export & Backup**: Export all scripts and run analytics directly to Markdown or JSON.
- **XDG Compliance**: Respects `$XDG_CONFIG_HOME`, `$XDG_DATA_HOME`, and `$XDG_STATE_HOME`.

---

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from Source

```bash
git clone https://github.com/jurekzsl/vellum.git
cd vellum
make build
# Optional: install to PATH
sudo mv vellum /usr/local/bin/
```

---

## Usage

### Interactive TUI Mode

Launch the full interactive terminal dashboard:

```bash
vellum
```

#### TUI Keybindings

- **Navigation**: `↑` / `↓` (or `k` / `j`)
- **Filter / Search**: `/` (Search by keyword or tag: `#s` for scripts, `#a` for aliases)
- **`enter`**: Run selected script or alias
- **`v`**: Open integrated **Log Viewer** to browse execution history
- **`r`**: Advanced run (custom parameters, directory path, non-blocking delay, copy output)
- **`a`**: Add a new script
- **`i`**: Import an existing script file
- **`l`**: Add a new shell alias
- **`e`**: Edit selected script / alias
- **`d`**: Delete selected script / alias
- **`c`**: Copy script or alias command to clipboard
- **`x`**: Export all scripts and metadata to Markdown
- **`o`**: Open script folder in native file manager
- **`s`**: Cycle sort order (`name` ➔ `type` ➔ `lastrun`)
- **`q`** / `ctrl+c`: Quit

#### Log Viewer Keybindings (when inside `v` mode)

- **`↑` / `↓` / `pgup` / `pgdn`**: Scroll log content
- **`n` / `p`** (or `tab` / `shift+tab`): View next (older) / previous (newer) log file
- **`c`**: Copy current log output to system clipboard
- **`g` / `G`**: Jump to top / bottom
- **`esc` / `q`**: Return to list view

---

### Headless CLI Mode

Run scripts directly from the terminal or in CI/CD pipelines:

```bash
# List all configured scripts and aliases
vellum list
vellum list --json

# Execute a script or alias headlessly with arguments
vellum run <name> [arguments...]

# Export scripts and metadata to stdout
vellum export md
vellum export json

# Print version
vellum version
```

---

## Configuration

Configuration is stored in `~/.config/vellum/config.json` (or `$XDG_CONFIG_HOME/vellum/config.json`).

```json
{
  "theme": "#BD93F9",
  "log_retention_days": 7,
  "max_log_files": 50,
  "default_shell": "/bin/zsh",
  "default_editor": "nano",
  "file_manager": "open",
  "sort_order": "name"
}
```

| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `theme` | String | `"#BD93F9"` | Hex color code for the TUI theme |
| `default_shell` | String | Auto (`$SHELL`) | Shell used to execute commands and wrappers |
| `default_editor` | String | Auto (`$EDITOR`) | External text editor |
| `file_manager` | String | Auto (`open` on macOS, `xdg-open` on Linux, `explorer` on Windows) | File manager utility |
| `sort_order` | String | `"name"` | Default sorting (`name`, `lastrun`, `type`) |
| `log_retention_days` | Integer | `7` | Days to keep log files before auto-deletion |
| `max_log_files` | Integer | `50` | Maximum log files per script before auto-cleanup |

---

## Supported Script Types

Vellum natively manages, detects, and executes scripts across 22 formats with language-specific boilerplates, shebangs, and interpreters:

| Category | Extensions | Default Interpreter / Command |
| :--- | :--- | :--- |
| **Coding** | `.py` | `python3` |
| | `.js` | `node` |
| | `.ts` | `ts-node` / `bun` / `deno` / `node` |
| | `.jsx` | `node` / `bun` |
| | `.tsx` | `ts-node` / `bun` |
| | `.rb` | `ruby` |
| | `.pl` | `perl` |
| | `.php` | `php` |
| | `.lua` | `lua` |
| | `.tcl` | `tclsh` / `tcl` |
| **Shell** | `.sh`, `.bash` | `bash` |
| | `.zsh` | `zsh` |
| | `.fish` | `fish` |
| | `.ps1` | `pwsh` / `powershell` |
| **Compiled (Source)** | `.go` | `go run` |
| | `.rs` | `rust-script` / `cargo run` |
| | `.java` | `java` (Single-file source execution) |
| | `.kt` | `kotlinc -script` / `kotlin` |
| | `.swift` | `swift` |
| **System** | `.applescript` | `osascript` |
| | `.bat`, `.cmd` | `cmd.exe /c` |
| | `.vbs` | `cscript //nologo` |

---

## License

MIT


