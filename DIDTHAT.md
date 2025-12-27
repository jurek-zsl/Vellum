# Vellum Project - Implementation Log

This document tracks the development progress and implemented features for the Vellum CLI/TUI tool.

## Core Setup
- [x] **Project Initialization**: Created Go module `github.com/jurekzsl/vellum` and directory structure.
- [x] **Dependencies**: Installed `bubbletea`, `bubbles`, `lipgloss`, and `huh`.
- [x] **Data Model**: Defined `Metadata` struct for Scripts and Aliases in `internal/model`.
- [x] **Configuration**: Implemented `config` package to manage persistent storage path (`~/.config/vellum/config.json`).
- [x] **Storage Engine**:
    - Created `store` package to handle file system operations.
    - Implemented `SaveScript` (creates subfolders) and `SaveAlias`.
    - Implemented `ListItems` to traverse directories and load metadata.
    - Implemented `DeleteItem` for removal.
    - Implemented `EnsureDirs` to create default directory structure (`~/vellum/scripts`, `~/vellum/aliases`).

## Features
- [x] **Script Management**:
    - Support for creating scripts via TUI Form (`a` key).
    - Support for importing existing scripts (`i` key).
    - Support for editing scripts (`e` key).
- [x] **Alias Management**:
    - Support for creating aliases (`l` key).
    - **Shell Integration**: Automatically appends new aliases to `.zshrc` (or detected shell RC file).
- [x] **Execution**:
    - Implemented `runner` package to execute scripts and aliases.
    - **Sudo Enforcement**: All scripts/aliases now run with `sudo` by default.
    - **Terminal Interaction**: Execution pauses terminal output ("Press Enter to return...") to allow viewing results.
    - **Logging**: Automatic logging of stdout/stderr to timestamped `.log` files in item directories.
    - **Log Cleanup**: Automatic deletion of logs older than 7 days on startup.
- [x] **Configuration**:
    - UI form to change Vellum storage directory (`c` key).

## TUI (Terminal User Interface)
- [x] **Layout**:
    - "Header - List - Footer" structure.
    - ASCII Logo in header.
    - **Boxed Design**: Main list is wrapped in a rounded border.
- [x] **List View**:
    - **Filtering**: Type to filter.
    - **Custom Filtering**: Added symbols `[s]` (script) and `[l]` (alias) to titles.
    - **Search Syntax**: Enables searching by `*s` or `*l` to filter by type.
    - Disabled default list title ("Vellum Scripts & Aliases") for cleaner look.
- [x] **Feedback & Status**:
    - Status bar integration for error reporting and success messages.
    - Persistent Footer showing all keybindings (`a`, `i`, `l`, `e`, `d`, `c`, `enter`, `h`, `q`).
- [x] **Forms**:
    - Utilized `huh` library for interactive input forms (Create, Import, Alias, Config, Edit).

## Bug Fixes
- [x] **Form Data Handling**: Fixed a critical bug where `huh` forms were not updating the model due to value receiver semantics. Changed `formData` to a pointer to persist updates across Bubble Tea loops.
- [x] **Import Logic**: Fixed silent failure on imports by adding error handling for file reading.
- [x] **Go Run Support**: Fixed execution of `go run` commands by correctly splitting arguments.
- [x] **Layout Sizing**: Fixed list height calculations to account for logo, footer, and border margins.
- [x] **Duplicate Messages**: Removed redundant "Action completed" and title messages to clean up the UI.
