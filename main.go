package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const rootPath = "/"

var (
	// Styles - Matching Textual's default dark theme roughly
	// Header: bright text on dark background
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1)

	// Footer: bright text on dark background
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Padding(0, 1)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7dd3fc")).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("#333333"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fde68a")).
			Bold(true)

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fde68a")).
			Bold(true).
			Underline(true)

	// Panels
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(0, 1)

	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7dd3fc"))

	// Buttons
	buttonStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 2).
			MarginRight(1)

	buttonFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#10b981")).
				Padding(0, 2).
				MarginRight(1).
				Foreground(lipgloss.Color("#10b981")).
				Bold(true)

	// Misc
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#a5b4fc"))
	summaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	loadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true).Align(lipgloss.Center)

	spinnerFrames = []string{"⠏", "⠛", "⠖", "⠒", "⠐", "⠐", "⠒", "⠖", "⠛"}
)

var columnSpecs = []ColumnSpec{
	{
		Key:       "number",
		Label:     "#",
		Width:     6,
		Accessor:  func(s Snapshot) string { return strconv.Itoa(s.Number) },
		SortField: "number",
	},
	{
		Key:       "snapshot_type",
		Label:     "Type",
		Width:     8,
		Accessor:  func(s Snapshot) string { return s.SnapshotType },
		SortField: "snapshot_type",
	},
	{
		Key:       "pre_number",
		Label:     "Pre",
		Width:     6,
		Accessor:  func(s Snapshot) string { return nullableInt(s.PreNumber) },
		SortField: "pre_number",
	},
	{
		Key:       "post_number",
		Label:     "Post",
		Width:     6,
		Accessor:  func(s Snapshot) string { return nullableInt(s.PostNumber) },
		SortField: "post_number",
	},
	{
		Key:       "date",
		Label:     "Date",
		Width:     20,
		Accessor:  func(s Snapshot) string { return s.Date },
		SortField: "date",
	},
	{
		Key:       "user",
		Label:     "User",
		Width:     10,
		Accessor:  func(s Snapshot) string { return s.User },
		SortField: "user",
	},
	{
		Key:       "cleanup",
		Label:     "Cleanup",
		Width:     10,
		Accessor:  func(s Snapshot) string { return s.Cleanup },
		SortField: "cleanup",
	},
	{
		Key:       "description",
		Label:     "Description",
		Width:     36,
		Accessor:  func(s Snapshot) string { return s.Description },
		SortField: "description",
	},
	{
		Key:       "used_space",
		Label:     "Size",
		Width:     12,
		Accessor:  func(s Snapshot) string { return humanReadableBytes(s.UsedSpace) },
		SortField: "used_space",
	},
	{
		Key:       "userdata",
		Label:     "Userdata",
		Width:     20,
		Accessor:  func(s Snapshot) string { return flattenUserData(s.Userdata) },
		SortField: "userdata",
	},
}

var sampleSnapshots = []Snapshot{
	{
		Config:       "root",
		Subvolume:    "/",
		Number:       0,
		SnapshotType: "single",
		Date:         "2025-11-18 00:00:00",
		User:         "root",
		Cleanup:      "number",
		Description:  "Initial baseline",
		UsedSpace:    ptrInt64(15 * 1024 * 1024 * 1024),
		Userdata:     map[string]string{"user": "admin"},
	},
	{
		Config:       "root",
		Subvolume:    "/",
		Number:       17,
		SnapshotType: "pre",
		PreNumber:    toOptionalInt(16),
		PostNumber:   toOptionalInt(18),
		Date:         "2025-11-19 06:12:34",
		User:         "root",
		Cleanup:      "number",
		Description:  "Before package update",
		UsedSpace:    ptrInt64(18 * 1024 * 1024 * 1024),
		Userdata:     map[string]string{"note": "patch"},
	},
	{
		Config:       "root",
		Subvolume:    "/",
		Number:       18,
		SnapshotType: "single",
		PreNumber:    toOptionalInt(17),
		Date:         "2025-11-19 12:45:10",
		User:         "root",
		Cleanup:      "number",
		Description:  "Midday checkpoint",
		UsedSpace:    ptrInt64(19 * 1024 * 1024 * 1024),
		Userdata:     map[string]string{"phase": "test"},
	},
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("snapper-TUI failed: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() UIState {
	m := UIState{
		Snapshots:         sampleSnapshots,
		Placeholder:       true,
		DetailOpen:        true,
		ActionMessage:     "Select a snapshot to preview the snapper commands.",
		Status:            "Loading snapshots...",
		Summary:           "Snapshots: 0 | Total used: 0 B | Free on /: ...",
		SortKey:           columnSpecs[0].SortField,
		SortReverse:       false,
		Loading:           true,
		SelectedSnapshots: make(map[int]bool),
		FocusedElement:    "table",
		ButtonRects:       make(map[string]Rect),
	}

	return m
}

func (m UIState) Init() tea.Cmd {
	return tea.Batch(refreshSnapshotsCmd(), tickCmd())
}

func (m UIState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TermWidth = msg.Width
		m.TermHeight = msg.Height

		// Calculate viewport height:
		// Header: 1
		// Table Header: 2 (Text + Border)
		// Scrollbar: 1
		// Action Panel: 6 (4 lines content + 2 border)
		// Summary: 1
		// Footer: 1
		// Total Fixed Overhead: 1 + 2 + 1 + 6 + 1 + 1 = 12 lines.
		// We add 1 extra line of buffer to be safe.
		availableHeight := msg.Height - 13
		if availableHeight < 1 {
			availableHeight = 1
		}
		m.ViewportHeight = availableHeight
		m.ensureCursorVisible()
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case TickMsg:
		return m.handleTick()
	case RefreshTriggerMsg:
		return m.handleRefreshTrigger()
	case RefreshResultMsg:
		return m.handleRefreshResult(RefreshResult(msg))
	case ActionResultMsg:
		return m.handleActionResult(ActionResult(msg))
	}
	return m, nil
}

func (m UIState) handleTick() (tea.Model, tea.Cmd) {
	if !m.Loading {
		return m, nil
	}
	m.SpinnerIndex = (m.SpinnerIndex + 1) % len(spinnerFrames)
	m.Status = fmt.Sprintf("%s Fetching snapshots from snapper...", spinnerFrames[m.SpinnerIndex])
	return m, tickCmd()
}

func (m UIState) handleRefreshTrigger() (tea.Model, tea.Cmd) {
	if m.Loading {
		return m, nil
	}
	m.Loading = true
	m.Status = "Refreshing snapshots..."
	return m, tea.Batch(refreshSnapshotsCmd(), tickCmd())
}

func (m UIState) handleRefreshResult(msg RefreshResult) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.Snapshots = sampleSnapshots
		m.Placeholder = true
		m.Status = fmt.Sprintf("snapper list failed: %v", msg.Err)
		m.Summary = buildSummary(m.Snapshots)
		m.ActionMessage = "Using sample data; install snapper for real snapshots."
		return m, nil
	}

	m.Snapshots = msg.Snapshots
	m.Placeholder = false
	m.sortSnapshots()
	if m.Cursor >= len(m.Snapshots) {
		m.Cursor = max(0, len(m.Snapshots)-1)
	}
	m.Summary = buildSummary(m.Snapshots)
	m.Status = fmt.Sprintf("Loaded %d snapshots", len(m.Snapshots))
	m.setActionPreview()
	return m, nil
}

func (m UIState) handleActionResult(msg ActionResult) (tea.Model, tea.Cmd) {
	m.ActionInProgress = false
	switch msg.Kind {
	case ActionDelete:
		if msg.Err == nil {
			m.ActionMessage = fmt.Sprintf("Snapshot(s) deleted successfully. Refreshing list...")
			m.Status = fmt.Sprintf("Deleted snapshot(s)")
			// Clear selection after successful delete
			m.SelectedSnapshots = make(map[int]bool)
			return m, waitRefreshCmd(time.Second)
		}
		m.ActionMessage = fmt.Sprintf("Delete failed: %s", msg.Output)
		m.Status = "Delete failed"
	case ActionRestore:
		if msg.Err == nil {
			m.ActionMessage = fmt.Sprintf("Rollback to snapshot %d complete. Press r to refresh or reboot.", msg.Snap.Number)
			m.Status = fmt.Sprintf("Applied snapshot %d", msg.Snap.Number)
			return m, nil
		}
		m.ActionMessage = fmt.Sprintf("Apply failed: %s", msg.Output)
		m.Status = "Apply failed"
	case ActionStatus:
		if msg.Err == nil {
			m.ActionMessage = fmt.Sprintf("Status output:\n%s", msg.Output)
			m.Status = "Status fetched"
			return m, nil
		}
		m.ActionMessage = fmt.Sprintf("Status failed: %s", msg.Output)
		m.Status = "Status failed"
	}
	return m, nil
}

func (m UIState) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.MouseWheelUp:
		if len(m.Snapshots) > 0 {
			m.Cursor = max(0, m.Cursor-1)
			m.ensureCursorVisible()
			m.setActionPreview()
		}
	case tea.MouseWheelDown:
		if len(m.Snapshots) > 0 {
			m.Cursor = min(len(m.Snapshots)-1, m.Cursor+1)
			m.ensureCursorVisible()
			m.setActionPreview()
		}
	case tea.MouseLeft:
		// Table area: typically left ~65% of screen
		tableWidth := int(float64(m.TermWidth) * 0.65)
		if tableWidth < 60 {
			tableWidth = 60
		}

		// Check if click is on table area (left panel)
		if msg.X < tableWidth {
			// Layout:
			// Y=0: App Header
			// Y=1: Table Header Text
			// Y=2: Table Header Border
			// Y=3: Row 0
			if msg.Y >= 1 && msg.Y <= 2 {
				// Calculate clicked column
				// Offset for "📋 " is 3 cells (2 for emoji + 1 space)
				x := msg.X - 3
				if x < 0 {
					return m, nil // Clicked on the icon
				}

				currentX := 0
				for _, spec := range columnSpecs {
					// Width + 1 for space separator
					w := spec.Width
					// Check if click is within this column's width
					if x >= currentX && x < currentX+w {
						m.toggleSort(spec.SortField)
						return m, nil
					}
					currentX += w + 1
				}
				return m, nil
			}

			// Table rows start at y=3
			if msg.Y >= 3 && msg.Y < m.TermHeight-4 {
				m.FocusedElement = "table"
				rowClicked := msg.Y - 3
				actualRowIdx := m.Offset + rowClicked
				if actualRowIdx < len(m.Snapshots) {
					m.Cursor = actualRowIdx
					m.ensureCursorVisible()
					m.setActionPreview()
				}
			}
		} else {
			// Right panel area
			// Calculate button positions dynamically
			restoreY, deleteY, statusY := m.getButtonYPositions()

			// Check X coordinate (right panel starts at tableWidth + 2)
			// Buttons are left-aligned in the panel, width ~16-20 chars
			panelStartX := tableWidth + 2
			if msg.X >= panelStartX && msg.X <= panelStartX+25 {
				if msg.Y >= restoreY && msg.Y < restoreY+3 {
					if !m.ActionInProgress && m.currentSnapshot() != nil {
						m.FocusedElement = "restore"
						m.ActionInProgress = true
						m.ActionMessage = "⏳ Executing apply..."
						return m, executeActionCmd(ActionRestore, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
					}
				}
				if msg.Y >= deleteY && msg.Y < deleteY+3 {
					if !m.ActionInProgress && m.currentSnapshot() != nil {
						m.FocusedElement = "delete"
						m.ActionInProgress = true
						m.ActionMessage = "⏳ Executing delete..."
						return m, executeActionCmd(ActionDelete, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
					}
				}
				if msg.Y >= statusY && msg.Y < statusY+3 {
					if !m.ActionInProgress && m.currentSnapshot() != nil {
						m.FocusedElement = "status"
						m.ActionInProgress = true
						m.ActionMessage = "⏳ Fetching status..."
						return m, executeActionCmd(ActionStatus, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
					}
				}
			}
		}
	}
	return m, nil
}

func (m UIState) getButtonYPositions() (int, int, int) {
	// Base Y offset for main content is 1 (Header takes 1 line)
	y := 1

	if m.DetailOpen {
		y += 1 // "Details" header
		y += 1 // Spacer

		if m.currentSnapshot() != nil {
			y += 11 // Content lines
			y += 2  // Border
		} else {
			y += 1 // "Select a snapshot..."
			y += 2 // Border
		}

		y += 2 // Spacer "\n\n"
	}

	y += 1 // "Actions" header
	y += 1 // Spacer

	restoreY := y
	y += 3 // Button height (1 text + 2 border)
	y += 1 // Spacer

	deleteY := y
	y += 3
	y += 1

	statusY := y

	return restoreY, deleteY, statusY
}

func (m UIState) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cmd := tea.Cmd(nil)

	// Global keys
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		// Cycle through focused elements
		elements := []string{"table", "restore", "delete", "status"}
		for i, e := range elements {
			if e == m.FocusedElement {
				m.FocusedElement = elements[(i+1)%len(elements)]
				return m, nil
			}
		}
		m.FocusedElement = "table"
		return m, nil
	}

	// Element-specific key handling
	switch m.FocusedElement {
	case "table":
		switch msg.String() {
		case "r":
			if !m.Loading {
				m.Loading = true
				m.Status = "Refreshing snapshots..."
				cmd = tea.Batch(refreshSnapshotsCmd(), tickCmd())
			}
		case "j", "down":
			if len(m.Snapshots) > 0 {
				m.Cursor = (m.Cursor + 1) % len(m.Snapshots)
				m.ensureCursorVisible()
				m.setActionPreview()
			}
		case "k", "up":
			if len(m.Snapshots) > 0 {
				m.Cursor = (m.Cursor - 1 + len(m.Snapshots)) % len(m.Snapshots)
				m.ensureCursorVisible()
				m.setActionPreview()
			}
		case "pgdn", "pagedown":
			if len(m.Snapshots) > 0 {
				m.Cursor = min(m.Cursor+m.ViewportHeight, len(m.Snapshots)-1)
				m.ensureCursorVisible()
				m.setActionPreview()
			}
		case "pgup", "pageup":
			if len(m.Snapshots) > 0 {
				m.Cursor = max(0, m.Cursor-m.ViewportHeight)
				m.ensureCursorVisible()
				m.setActionPreview()
			}
		case "enter":
			m.DetailOpen = !m.DetailOpen
		case " ":
			if len(m.Snapshots) > 0 {
				snap := m.Snapshots[m.Cursor]
				if m.SelectedSnapshots[snap.Number] {
					delete(m.SelectedSnapshots, snap.Number)
				} else {
					m.SelectedSnapshots[snap.Number] = true
				}
				m.setActionPreview()
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			m.updateSortKey(msg.String())
		}

	case "restore":
		if msg.String() == "enter" {
			if !m.ActionInProgress && m.currentSnapshot() != nil {
				m.ActionInProgress = true
				m.ActionMessage = "⏳ Executing apply..."
				cmd = executeActionCmd(ActionRestore, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
			}
		}

	case "delete":
		if msg.String() == "enter" {
			if !m.ActionInProgress && m.currentSnapshot() != nil {
				m.ActionInProgress = true
				m.ActionMessage = "⏳ Executing delete..."
				cmd = executeActionCmd(ActionDelete, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
			}
		}

	case "status":
		if msg.String() == "enter" {
			if !m.ActionInProgress && m.currentSnapshot() != nil {
				m.ActionInProgress = true
				m.ActionMessage = "⏳ Fetching status..."
				cmd = executeActionCmd(ActionStatus, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
			}
		}
	}

	// Keyboard shortcuts available in all modes
	if m.FocusedElement == "table" {
		switch msg.String() {
		case "A", "a":
			if !m.ActionInProgress && m.currentSnapshot() != nil {
				m.ActionInProgress = true
				m.ActionMessage = "⏳ Executing apply..."
				cmd = executeActionCmd(ActionRestore, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
			}
		case "D", "d":
			if !m.ActionInProgress && m.currentSnapshot() != nil {
				m.ActionInProgress = true
				m.ActionMessage = "⏳ Executing delete..."
				cmd = executeActionCmd(ActionDelete, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
			}
		case "s":
			if !m.ActionInProgress && m.currentSnapshot() != nil {
				m.ActionInProgress = true
				m.ActionMessage = "⏳ Fetching status..."
				cmd = executeActionCmd(ActionStatus, *m.currentSnapshot(), m.SelectedSnapshots, m.Snapshots)
			}
		}
	}

	return m, cmd
}

func (m *UIState) updateSortKey(key string) {
	index := keyToColumnIndex(key)
	if index < 0 {
		return
	}
	spec := columnSpecs[index]
	m.toggleSort(spec.SortField)
}

func (m *UIState) toggleSort(field string) {
	if m.SortKey == field {
		m.SortReverse = !m.SortReverse
	} else {
		m.SortKey = field
		m.SortReverse = false
	}
	m.sortSnapshots()
	m.ensureCursorVisible()
	m.setActionPreview()
	m.Status = fmt.Sprintf("Sorting by %s", m.SortKey)
}

func keyToColumnIndex(key string) int {
	if key == "0" {
		return 9
	}
	n, err := strconv.Atoi(key)
	if err != nil {
		return -1
	}
	return n - 1
}

func (m *UIState) sortSnapshots() {
	sort.SliceStable(m.Snapshots, func(i, j int) bool {
		iVal := valueForSort(m.Snapshots[i], m.SortKey)
		jVal := valueForSort(m.Snapshots[j], m.SortKey)
		if m.SortReverse {
			return jVal < iVal
		}
		return iVal < jVal
	})
}

func (m *UIState) currentSnapshot() *Snapshot {
	if len(m.Snapshots) == 0 {
		return nil
	}
	if m.Cursor >= len(m.Snapshots) {
		m.Cursor = len(m.Snapshots) - 1
	}
	return &m.Snapshots[m.Cursor]
}

func (m *UIState) ensureCursorVisible() {
	if m.ViewportHeight <= 0 {
		m.ViewportHeight = 10 // fallback
	}

	// If cursor is above viewport, scroll up
	if m.Cursor < m.Offset {
		m.Offset = m.Cursor
	}

	// If cursor is below viewport, scroll down
	if m.Cursor >= m.Offset+m.ViewportHeight {
		m.Offset = m.Cursor - m.ViewportHeight + 1
	}

	// Clamp offset to valid range
	if m.Offset < 0 {
		m.Offset = 0
	}
	if m.Offset > max(0, len(m.Snapshots)-m.ViewportHeight) {
		m.Offset = max(0, len(m.Snapshots)-m.ViewportHeight)
	}
}

func (m *UIState) setActionPreview() {
	if snap := m.currentSnapshot(); snap != nil {
		start := computeStatusStart(*snap)
		m.ActionMessage = strings.Join([]string{
			fmt.Sprintf("Apply: sudo snapper rollback %d", snap.Number),
			fmt.Sprintf("Delete: sudo snapper delete %d", snap.Number),
			fmt.Sprintf("Status: sudo snapper status %d..%d", start, snap.Number),
			"[A]pply • [D]elete • [S]tatus • Click buttons or press Tab+Enter",
		}, "\n")
		return
	}
	m.ActionMessage = "Select a snapshot to preview the snapper commands."
}

func (m UIState) View() string {
	// Handle initial state where dimensions might be 0
	width := m.TermWidth
	if width == 0 {
		width = 80
	}
	height := m.TermHeight
	if height == 0 {
		height = 24
	}

	// 1. Header
	header := headerStyle.Width(width).Render("Snapper TUI")

	// 2. Main content
	var mainContent string
	if m.Loading {
		loadingText := fmt.Sprintf("%s Loading snapshots...", spinnerFrames[m.SpinnerIndex])
		// Center the loading text in the available space
		// Available height = height - header(1) - action(3) - summary(1) - footer(1) = height - 6
		availH := max(5, height-6)
		mainContent = lipgloss.Place(width, availH, lipgloss.Center, lipgloss.Center, loadingStyle.Render(loadingText))
	} else {
		// Calculate max height for content
		// Table height = ViewportHeight + 2 (Header) + 1 (Scrollbar) = ViewportHeight + 3
		// We want right panel to match this roughly
		maxContentHeight := m.ViewportHeight + 3

		tableView := m.renderTable()
		rightPanel := m.renderRightPanel(maxContentHeight)

		// Use JoinHorizontal to combine them safely
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, tableView, "  ", rightPanel)
	}

	// 3. Action message
	actionMsg := panelStyle.Width(width - 2).Render(m.ActionMessage)

	// 4. Summary
	summary := summaryStyle.Render(m.Summary)

	// 5. Footer
	footerText := "q: Quit | r: Refresh | Tab: Navigate | Enter: Select"
	footer := footerStyle.Width(width).Render(footerText)

	// Combine all parts vertically
	// We use JoinVertical to ensure everything is stacked correctly
	ui := lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainContent,
		actionMsg,
		summary,
		footer,
	)

	// Force the final output to fit the terminal size exactly
	// This prevents "cutting off" by ensuring we don't emit more lines than the terminal has
	return lipgloss.Place(width, height, lipgloss.Top, lipgloss.Left, ui)
}

func (m UIState) renderTable() string {
	var b strings.Builder

	// Calculate table width
	width := m.TermWidth
	if width == 0 {
		width = 80
	}
	tableWidth := int(float64(width) * 0.65)
	if tableWidth < 60 {
		tableWidth = 60
	}

	// Header: "📋 " + column labels
	var headerCells []string
	for _, spec := range columnSpecs {
		headerCells = append(headerCells, padOrTruncate(spec.Label, spec.Width))
	}
	headerLine := "📋 " + strings.Join(headerCells, " ")

	// Ensure header fits within tableWidth
	if len(headerLine) > tableWidth {
		headerLine = headerLine[:tableWidth-1] + "…"
	}
	// Pad header to fill full width
	headerLine = padOrTruncate(headerLine, tableWidth)

	if m.FocusedElement == "table" {
		b.WriteString(tableHeaderStyle.Render(headerLine + " (↑↓/PgUp/PgDn)"))
	} else {
		b.WriteString(tableHeaderStyle.Render(headerLine))
	}
	b.WriteString("\n")

	if len(m.Snapshots) == 0 {
		emptyMsg := "No snapshots available. Press r to try again."
		emptyMsg = padOrTruncate(emptyMsg, tableWidth)
		b.WriteString(emptyMsg)
		b.WriteString("\n")
		return b.String()
	}

	// Determine viewport range
	endIdx := min(m.Offset+m.ViewportHeight, len(m.Snapshots))

	// Show rows only in viewport
	for idx := m.Offset; idx < endIdx; idx++ {
		snap := m.Snapshots[idx]

		// Row style
		var rowStyle lipgloss.Style
		isSelected := m.SelectedSnapshots[snap.Number]

		if idx == m.Cursor && m.FocusedElement == "table" {
			rowStyle = focusedStyle
		} else if isSelected {
			rowStyle = selectedStyle
		} else {
			rowStyle = lipgloss.NewStyle()
		}

		// Render columns
		var rowCells []string
		for colIdx, spec := range columnSpecs {
			val := spec.Accessor(snap)

			// Add selection indicator to the first column
			if colIdx == 0 && isSelected {
				val = "✔ " + val
			}

			// Pad/Truncate
			cell := padOrTruncate(val, spec.Width)
			rowCells = append(rowCells, cell)
		}
		line := strings.Join(rowCells, " ")

		// Truncate if too long
		if len(line) > tableWidth {
			line = line[:tableWidth-1] + "…"
		}

		// Pad to fill full width - CRITICAL for preventing bleed
		line = padOrTruncate(line, tableWidth)

		// Apply style to the full padded line
		line = rowStyle.Render(line)

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Show scrollbar indicator
	if len(m.Snapshots) > m.ViewportHeight {
		scrollPercent := int((float64(m.Offset) / float64(len(m.Snapshots)-m.ViewportHeight)) * 100)
		scrollLine := fmt.Sprintf("Scroll: %d%% [%d-%d of %d]", scrollPercent, m.Offset+1, endIdx, len(m.Snapshots))
		scrollLine = padOrTruncate(scrollLine, tableWidth)
		b.WriteString(scrollLine)
		b.WriteString("\n")
	}

	return b.String()
}

func (m UIState) renderRightPanel(maxHeight int) string {
	// Calculate right panel width
	width := m.TermWidth
	if width == 0 {
		width = 80
	}

	tableWidth := int(float64(width) * 0.65)
	if tableWidth < 60 {
		tableWidth = 60
	}
	rightPanelWidth := width - tableWidth - 2
	if rightPanelWidth < 20 {
		rightPanelWidth = 20
	}

	// Calculate fixed height requirements for Actions section
	// "Actions" header (1) + Spacer (1) + Restore (3) + Spacer (1) + Delete (3) + Spacer (1) + Status (3) = 13 lines
	// Let's verify:
	// Header: 1
	// Spacer: 1
	// Restore: 3 (1 text + 2 border)
	// Spacer: 1
	// Delete: 3
	// Spacer: 1
	// Status: 3
	// Total Actions Height = 13
	actionsHeight := 13

	// Available height for details
	detailsAvailableHeight := maxHeight - actionsHeight
	if detailsAvailableHeight < 0 {
		detailsAvailableHeight = 0
	}

	var b strings.Builder

	// Details section
	if m.DetailOpen && detailsAvailableHeight > 2 { // Need at least some space
		var detailsBuilder strings.Builder
		detailsBuilder.WriteString(detailHeaderStyle.Render("Details"))
		detailsBuilder.WriteString("\n")

		contentHeight := detailsAvailableHeight - 2 // -1 for Header, -1 for spacer/margin
		if contentHeight > 0 {
			if snap := m.currentSnapshot(); snap != nil {
				detailLines := []string{
					fmt.Sprintf("Config: %s", snap.Config),
					fmt.Sprintf("Subvolume: %s", snap.Subvolume),
					fmt.Sprintf("Number: %d (%s)", snap.Number, snap.SnapshotType),
					fmt.Sprintf("Desc: %s", nonEmpty(snap.Description, "<none>")),
					fmt.Sprintf("User: %s", snap.User),
					fmt.Sprintf("Cleanup: %s", nonEmpty(snap.Cleanup, "<none>")),
					fmt.Sprintf("Pre #: %s | Post #: %s", nullableInt(snap.PreNumber), nullableInt(snap.PostNumber)),
					fmt.Sprintf("Default: %s | Active: %s", boolText(snap.Default), boolText(snap.Active)),
					fmt.Sprintf("Size: %s", humanReadableBytes(snap.UsedSpace)),
					fmt.Sprintf("Data: %s", nonEmpty(flattenUserData(snap.Userdata), "<none>")),
				}

				// Truncate lines if they don't fit
				if len(detailLines) > contentHeight {
					detailLines = detailLines[:contentHeight]
				}

				// Account for border width (2 chars on each side = 4 total)
				contentWidth := rightPanelWidth - 4
				if contentWidth < 10 {
					contentWidth = 10
				}
				detailsBuilder.WriteString(panelStyle.Width(contentWidth).Render(strings.Join(detailLines, "\n")))
			} else {
				contentWidth := rightPanelWidth - 4
				if contentWidth < 10 {
					contentWidth = 10
				}
				detailsBuilder.WriteString(panelStyle.Width(contentWidth).Render("Select a snapshot to view details."))
			}
		}
		detailsBuilder.WriteString("\n\n") // Spacer

		// Only write details if we haven't exceeded height
		// Actually, we calculated detailsAvailableHeight to fit.
		b.WriteString(detailsBuilder.String())
	} else {
		// If details are closed or no space, fill with newlines to push actions down?
		// Or just render actions at the top?
		// Python version has actions at bottom? No, it has a vertical layout.
		// Let's just render actions.
		// If we want to align actions to bottom, we need to fill space.
		// But simpler to just stack them.
		// If we want to preserve layout stability, maybe fill space?
		// Let's just stack for now to avoid overflow.
		if m.DetailOpen {
			// If we wanted to show details but couldn't, maybe show a message?
			// b.WriteString("Details hidden (too small)\n\n")
		}
	}

	// Action buttons
	b.WriteString(detailHeaderStyle.Render("Actions"))
	b.WriteString("\n")

	restoreBtn := buttonStyle.Render("A: Apply")
	deleteBtn := buttonStyle.Render("D: Delete")
	statusBtn := buttonStyle.Render("S: Status")

	if m.FocusedElement == "restore" {
		restoreBtn = buttonFocusStyle.Render("A: Apply")
	} else if m.FocusedElement == "delete" {
		deleteBtn = buttonFocusStyle.Render("D: Delete")
	} else if m.FocusedElement == "status" {
		statusBtn = buttonFocusStyle.Render("S: Status")
	}

	b.WriteString(restoreBtn)
	b.WriteString("\n")
	b.WriteString(deleteBtn)
	b.WriteString("\n")
	b.WriteString(statusBtn)

	result := b.String()

	// Ensure the entire panel has explicit width
	// Use lipgloss to set width on the whole panel
	panelContainer := lipgloss.NewStyle().Width(rightPanelWidth)
	return panelContainer.Render(result)
}

// Command functions
func refreshSnapshotsCmd() tea.Cmd {
	return func() tea.Msg {
		snaps, err := listSnapshots()
		return RefreshResultMsg{Snapshots: snaps, Err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*120, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func waitRefreshCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return RefreshTriggerMsg{}
	})
}

func executeActionCmd(kind ActionKind, snap Snapshot, selected map[int]bool, allSnaps []Snapshot) tea.Cmd {
	return func() tea.Msg {
		// Determine targets
		var targets []Snapshot
		if len(selected) > 0 {
			// If selection exists, use it
			for _, s := range allSnaps {
				if selected[s.Number] {
					targets = append(targets, s)
				}
			}
		} else {
			// Otherwise use the single target
			targets = []Snapshot{snap}
		}

		// Validation
		if len(targets) > 1 {
			if kind == ActionRestore || kind == ActionStatus {
				return ActionResultMsg{
					Kind:   kind,
					Snap:   snap,
					Err:    fmt.Errorf("cannot perform %s on multiple snapshots", kind),
					Output: "Please select only one snapshot.",
				}
			}
		}

		// Execute
		if kind == ActionDelete {
			// Batch delete
			// Optimization: check if we can do batch delete with snapper
			// snapper delete 1 2 3
			var args []string
			args = append(args, "delete")
			for _, t := range targets {
				args = append(args, fmt.Sprint(t.Number))
			}

			cmd := exec.Command("snapper", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return ActionResultMsg{
					Kind:   kind,
					Snap:   snap, // Representative
					Err:    err,
					Output: string(output),
				}
			}
			return ActionResultMsg{
				Kind:   kind,
				Snap:   snap,
				Err:    nil,
				Output: "All selected snapshots deleted.",
			}
		}

		// Single action (Restore/Status)
		// Re-use existing logic for single action
		target := targets[0]
		var args []string
		switch kind {
		case ActionRestore:
			args = []string{"rollback", fmt.Sprint(target.Number)}
		case ActionStatus:
			start := computeStatusStart(target)
			args = []string{"status", fmt.Sprintf("%d..%d", start, target.Number)}
		}

		cmd := exec.Command("snapper", args...)
		output, err := cmd.CombinedOutput()
		trimmed := strings.TrimSpace(string(output))
		if len(trimmed) > 500 {
			trimmed = trimmed[:497] + "..."
		}
		return ActionResultMsg{Kind: kind, Snap: target, Output: trimmed, Err: err}
	}
}
