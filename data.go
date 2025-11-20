package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// listSnapshots executes snapper list command and returns parsed snapshots
func listSnapshots() ([]Snapshot, error) {
	columns := []string{
		"config",
		"subvolume",
		"number",
		"type",
		"pre-number",
		"post-number",
		"date",
		"user",
		"cleanup",
		"description",
		"userdata",
		"used-space",
		"default",
		"active",
	}

	cmd := exec.Command("snapper", "--jsonout", "list", "--columns", strings.Join(columns, ","))
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("snapper list failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("snapper list failed: %w", err)
	}

	var payload map[string][]map[string]interface{}
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("unable to decode snapper JSON: %w", err)
	}

	var snaps []Snapshot
	for configName, entries := range payload {
		for _, entry := range entries {
			snaps = append(snaps, snapshotFromRaw(configName, entry))
		}
	}

	return snaps, nil
}

// snapshotFromRaw converts a raw JSON map to a Snapshot struct
func snapshotFromRaw(config string, data map[string]interface{}) Snapshot {
	number, _ := toInt(data["number"])
	pre := toOptionalInt(data["pre-number"])
	post := toOptionalInt(data["post-number"])
	used, _ := toInt64(data["used-space"])
	userdata := toStringMap(data["userdata"])

	return Snapshot{
		Config:       toString(data["config"], config),
		Subvolume:    toString(data["subvolume"], ""),
		Number:       number,
		SnapshotType: toString(data["type"], ""),
		PreNumber:    pre,
		PostNumber:   post,
		Date:         toString(data["date"], ""),
		User:         toString(data["user"], ""),
		Cleanup:      toString(data["cleanup"], ""),
		Description:  toString(data["description"], ""),
		Userdata:     userdata,
		UsedSpace:    used,
		Default:      toBool(data["default"]),
		Active:       toBool(data["active"]),
	}
}

// toString converts a value to string with a fallback
func toString(value interface{}, fallback string) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case nil:
		return fallback
	default:
		return fmt.Sprint(v)
	}
}

// toInt converts a value to int
func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}

// toInt64 converts a value to *int64
func toInt64(value interface{}) (*int64, bool) {
	switch v := value.(type) {
	case int:
		converted := int64(v)
		return &converted, true
	case float64:
		converted := int64(v)
		return &converted, true
	case int64:
		return &v, true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return &i, true
		}
	}
	return nil, false
}

// toOptionalInt converts a value to *int
func toOptionalInt(value interface{}) *int {
	if i, ok := toInt(value); ok {
		return &i
	}
	return nil
}

// toStringMap converts a value to map[string]string
func toStringMap(value interface{}) map[string]string {
	result := map[string]string{}
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			result[key] = fmt.Sprint(val)
		}
		return result
	case map[string]string:
		return v
	default:
		return nil
	}
}

// toBool converts a value to bool
func toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	case int:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}

// freeSpaceForPath returns the free disk space for a given path
func freeSpaceForPath(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}
