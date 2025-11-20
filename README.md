# snapper-TUI (Go Edition)

A high-performance terminal user interface (TUI) for managing Btrfs **snapper** snapshots, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) in Go. Browse, sort, filter, restore, and delete snapshots with an intuitive, responsive layout inspired by the Python Textual version.

## Demo

![Snapper TUI Demo](./snapper-tui-go-demo_github.gif)

## Features

### Core Functionality
- **Sortable Snapshot Table:** Browse all snapshot metadata with automatic scrolling
  - Only visible rows are rendered (smooth performance)
  - Cursor auto-scrolls viewport when navigating
  - Scroll indicator shows position and range
  - Page Up/Down for quick navigation
- **Column Sorting:**
  - Press number keys (`1`–`9`, `0`) to sort by any column
  - Repeat to toggle ascending/descending order
  - Click table headers with mouse to sort
- **Multi-Selection:** Select multiple snapshots for batch operations
  - Press `space` to toggle selection on current snapshot
  - Selected snapshots are highlighted in the table
  - Perform batch deletes on multiple snapshots
- **Detailed Preview Panel:** Right-side panel shows full snapshot metadata with a clean, organized layout
  - Toggle visibility with `enter` key
  - Shows comprehensive snapshot information in an organized format
- **Interactive Action Buttons:**
  - Visual buttons with focus highlighting
  - Click with mouse or Tab+Enter to activate
  - Real-time command preview before execution
- **Keyboard Shortcuts:** Direct command execution with quick keys
  - Press `A`/`a` to apply/restore selected snapshot
  - Press `D`/`d` to delete selected snapshot(s)
  - Press `s` to show status diff for snapshot range
- **Mouse Support:**
  - Click table rows to select
  - Click buttons to execute actions
  - Click headers to sort by column
  - Mouse wheel to scroll through snapshots
- **Animated Loading:** Smooth braille spinner while fetching snapshot data
- **Space Tracking:** Real-time disk usage (total used, free space, snapshot count)
- **Auto-refresh:** Snapshot list refreshes after successful deletion
- **Focus Navigation:** Tab/Shift+Tab between table and action buttons
- **Fallback Mode:** Works with sample data when `snapper` unavailable

## Installation

```bash
git clone https://github.com/ulolol/snapper-TUI-go.git
cd snapper-TUI-go
go build ./...
```

Ensure the `snapper` binary is installed and available in your `PATH`. Running the app without the command still works via sample data.

## Usage

```bash
sudo ./snapper-TUI
```

### Keybindings

#### Navigation & Focus
| Key | Action |
|-----|--------|
| `tab` | Cycle through: filter → table → restore button → delete button → status button → filter |
| `shift+tab` | Cycle backwards |
| `/` | Focus filter input (shortcut from table) |
| `enter` | Toggle detail panel (when in table) |

#### Table Navigation (when table is focused)
| Key | Action |
|-----|--------|
| `↑` / `↓` or `k` / `j` | Move cursor up/down (auto-scrolls viewport) |
| `PgUp` / `PgDn` | Scroll up/down by page |
| `space` | Toggle multi-selection for current snapshot |
| `1`–`9`, `0` | Sort by column (1=#, 2=Type, 3=Pre, 4=Post, 5=Date, 6=User, 7=Cleanup, 8=Desc, 9=Size, 0=Userdata) |
| `r` | Refresh snapshot list |
| `A` / `a` | Apply/Restore the selected snapshot (rollback) |
| `D` / `d` | Delete the selected snapshot(s) |
| `s` | Show status diff for snapshot range |



#### Button Activation (when button is focused)
| Key | Action |
|-----|--------|
| `enter` | Execute the focused action |

#### Global
| Key | Action |
|-----|--------|
| `q` or `Ctrl+C` | Quit the TUI |

### Mouse Support

- **Click table rows** to select a snapshot
- **Click action buttons** to execute directly (Apply, Delete, Status)
- **Click column headers** to sort by that column
- **Mouse wheel** to scroll up/down through snapshots

### Requirements

- Linux with the Btrfs filesystem and the `snapper` command installed
- Root privileges (use `sudo` when running the binary)

## Project Layout

```
snapper-TUI-go/
├── main.go             # Bubble Tea model, UI rendering, and command wiring
├── models.go           # Data structures (Snapshot, UIState, message types)
├── data.go             # Snapper CLI interaction and JSON parsing
├── utils.go            # Helper functions (formatting, sorting, calculations)
├── background.go       # Background image support and color utilities
├── go.mod / go.sum     # Go module dependencies (Bubble Tea, Lip Gloss, Imaging)
├── README.md           # Documentation (this file)
├── AILog.md            # Development progress and findings
└── LICENSE             # MIT License
```

### Architecture

The application is structured using clean separation of concerns:

- **models.go**: Defines all data structures and custom Bubble Tea message types
- **data.go**: Handles snapper CLI communication and data parsing from JSON
- **utils.go**: Provides utility functions for formatting and sorting
- **main.go**: Contains the Bubble Tea app logic, UI rendering, and state management

## License

[MIT](LICENSE)
