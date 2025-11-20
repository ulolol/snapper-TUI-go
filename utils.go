package main

import (
	"fmt"
	"sort"
	"strings"
)

// humanReadableBytes converts bytes to human-readable format
func humanReadableBytes(value *int64) string {
	if value == nil {
		return "n/a"
	}
	return humanReadableFromUint64(uint64(*value))
}

// humanReadableFromUint64 converts uint64 bytes to human-readable format
func humanReadableFromUint64(value uint64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	amount := float64(value)
	idx := 0
	for amount >= 1024 && idx < len(units)-1 {
		amount /= 1024
		idx++
	}
	return fmt.Sprintf("%.1f %s", amount, units[idx])
}

// flattenUserData converts userdata map to string
func flattenUserData(data map[string]string) string {
	if len(data) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(data))
	for key, val := range data {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, val))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, " ")
}

// buildSummary creates the summary line
func buildSummary(snaps []Snapshot) string {
	total := int64(0)
	for _, snap := range snaps {
		total += int64Value(snap.UsedSpace)
	}
	free, err := freeSpaceForPath(rootPath)
	usedText := humanReadableBytes(ptrInt64(total))
	freeText := "n/a"
	if err == nil {
		freeText = humanReadableFromUint64(free)
	}
	return fmt.Sprintf("Snapshots: %d | Total used: %s | Free on %s: %s", len(snaps), usedText, rootPath, freeText)
}

// padOrTruncate pads or truncates a string to a specific width
func padOrTruncate(value string, width int) string {
	if width <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) > width {
		return string(runes[:width-1]) + "…"
	}
	return fmt.Sprintf("%-*s", width, value)
}

// nullableInt converts *int to string
func nullableInt(value *int) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *value)
}

// intValue gets value from *int
func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

// int64Value gets value from *int64
func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

// ptrInt64 creates a pointer to int64
func ptrInt64(v int64) *int64 {
	return &v
}

// nonEmpty returns a string or a fallback if empty
func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// boolText converts bool to "yes"/"no"
func boolText(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

// max returns the maximum of two ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// computeStatusStart computes the start number for status command
func computeStatusStart(snap Snapshot) int {
	if snap.PreNumber != nil {
		return *snap.PreNumber
	}
	if snap.Number > 0 {
		return snap.Number - 1
	}
	return 0
}

// valueForSort gets the value to use for sorting
func valueForSort(s Snapshot, key string) string {
	switch key {
	case "number":
		return fmt.Sprintf("%08d", s.Number)
	case "pre_number":
		return fmt.Sprintf("%08d", intValue(s.PreNumber))
	case "post_number":
		return fmt.Sprintf("%08d", intValue(s.PostNumber))
	case "used_space":
		return fmt.Sprintf("%016d", int64Value(s.UsedSpace))
	case "userdata":
		return flattenUserData(s.Userdata)
	case "snapshot_type":
		return s.SnapshotType
	case "user":
		return s.User
	case "cleanup":
		return s.Cleanup
	case "description":
		return s.Description
	case "date":
		return s.Date
	default:
		return s.Date
	}
}

// getActionPreview returns the command preview for a snapshot
func getActionPreview(snap Snapshot) string {
	if snap.Number < 0 {
		return "Select a snapshot to preview the snapper commands."
	}
	start := computeStatusStart(snap)
	return strings.Join([]string{
		fmt.Sprintf("Restore: sudo snapper rollback %d", snap.Number),
		fmt.Sprintf("Delete: sudo snapper delete %d", snap.Number),
		fmt.Sprintf("Status: sudo snapper status %d..%d", start, snap.Number),
	}, "\n")
}
