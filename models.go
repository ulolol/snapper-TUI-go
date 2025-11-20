package main

import "time"

// Snapshot represents a snapper snapshot
type Snapshot struct {
	Config       string
	Subvolume    string
	Number       int
	SnapshotType string
	PreNumber    *int
	PostNumber   *int
	Date         string
	User         string
	Cleanup      string
	Description  string
	Userdata     map[string]string
	UsedSpace    *int64
	Default      bool
	Active       bool
}

// ColumnSpec defines a column in the snapshot table
type ColumnSpec struct {
	Key       string
	Label     string
	Width     int
	Accessor  func(Snapshot) string
	SortField string
}

// UIState represents the state of the application
type UIState struct {
	Snapshots         []Snapshot
	Cursor            int
	Offset            int // for scrolling
	SortKey           string
	SortReverse       bool
	Loading           bool
	SpinnerIndex      int
	ActionMessage     string
	Status            string
	Summary           string
	Placeholder       bool
	ActionInProgress  bool
	DetailOpen        bool
	SelectedSnapshot  *Snapshot
	SelectedSnapshots map[int]bool // Set of selected snapshot numbers
	FocusedElement    string       // "table", "restore", "delete", "status"
	TableRect         Rect
	ButtonRects       map[string]Rect // button ID -> rectangle
	TermWidth         int
	TermHeight        int
	ViewportHeight    int // how many rows fit on screen
}

// Rect represents a rectangular area for mouse tracking
type Rect struct {
	X, Y, Width, Height int
}

// ActionKind represents the type of action to execute
type ActionKind int

const (
	ActionUnknown ActionKind = iota
	ActionRestore
	ActionDelete
	ActionStatus
)

// ActionResult represents the result of an action
type ActionResult struct {
	Kind   ActionKind
	Snap   Snapshot
	Output string
	Err    error
}

// RefreshResult represents the result of a refresh operation
type RefreshResult struct {
	Snapshots []Snapshot
	Err       error
}

// Custom message types for Bubble Tea
type RefreshTriggerMsg struct{}
type TickMsg time.Time

type RefreshResultMsg struct {
	Snapshots []Snapshot
	Err       error
}

type ActionResultMsg struct {
	Kind   ActionKind
	Snap   Snapshot
	Output string
	Err    error
}

type MouseClickMsg struct {
	X, Y int
}

type WindowSizeMsg struct {
	Width, Height int
}
