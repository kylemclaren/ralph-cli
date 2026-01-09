// Package claudecode provides integration with Claude Code's hook system.
// This allows Ralph to coordinate with Claude Code hooks and expose Ralph state
// to hook scripts via environment variables.
package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings represents Claude Code's settings.json structure
type Settings struct {
	Hooks map[string][]HookMatcher `json:"hooks"`
}

// HookMatcher matches tools/events to hooks
type HookMatcher struct {
	Matcher string `json:"matcher"`
	Hooks   []Hook `json:"hooks"`
}

// Hook represents a single hook configuration
type Hook struct {
	Type    string `json:"type"`    // "command" or "prompt"
	Command string `json:"command"` // For type: "command"
	Prompt  string `json:"prompt"`  // For type: "prompt"
	Timeout int    `json:"timeout"` // Timeout in seconds (default 60)
}

// HookEvent types that Claude Code supports
const (
	EventPreToolUse        = "PreToolUse"
	EventPostToolUse       = "PostToolUse"
	EventPermissionRequest = "PermissionRequest"
	EventNotification      = "Notification"
	EventUserPromptSubmit  = "UserPromptSubmit"
	EventStop              = "Stop"
	EventSubagentStop      = "SubagentStop"
	EventPreCompact        = "PreCompact"
	EventSessionStart      = "SessionStart"
	EventSessionEnd        = "SessionEnd"
)

// SettingsPaths returns the paths to check for Claude Code settings
// in order of precedence (highest to lowest)
func SettingsPaths() []string {
	home, _ := os.UserHomeDir()

	paths := []string{
		// Project local settings (not in git)
		".claude/settings.local.json",
		// Project settings (in git)
		".claude/settings.json",
	}

	// User settings
	if home != "" {
		paths = append(paths, filepath.Join(home, ".claude", "settings.json"))
	}

	return paths
}

// LoadSettings loads Claude Code settings from the standard locations
func LoadSettings() (*Settings, error) {
	merged := &Settings{
		Hooks: make(map[string][]HookMatcher),
	}

	// Load in reverse precedence order so higher precedence overwrites
	paths := SettingsPaths()
	for i := len(paths) - 1; i >= 0; i-- {
		settings, err := LoadSettingsFromPath(paths[i])
		if err != nil {
			continue // File doesn't exist or invalid, skip
		}

		// Merge hooks
		for event, matchers := range settings.Hooks {
			merged.Hooks[event] = matchers
		}
	}

	return merged, nil
}

// LoadSettingsFromPath loads settings from a specific path
func LoadSettingsFromPath(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// HasHooks returns true if any hooks are configured for the given event
func (s *Settings) HasHooks(event string) bool {
	matchers, ok := s.Hooks[event]
	return ok && len(matchers) > 0
}

// GetHooks returns all hooks for a given event
func (s *Settings) GetHooks(event string) []HookMatcher {
	return s.Hooks[event]
}

// IsClaudeCodeAvailable checks if Claude Code CLI is available
func IsClaudeCodeAvailable() bool {
	// Check if 'claude' command exists
	paths := []string{
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude",
	}

	// Also check PATH
	pathEnv := os.Getenv("PATH")
	if pathEnv != "" {
		for _, dir := range filepath.SplitList(pathEnv) {
			paths = append(paths, filepath.Join(dir, "claude"))
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}

	return false
}

// SettingsExist checks if any Claude Code settings files exist
func SettingsExist() bool {
	for _, path := range SettingsPaths() {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}
