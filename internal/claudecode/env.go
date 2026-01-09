package claudecode

import (
	"fmt"
	"os"
	"strconv"
)

// Environment variable names that Ralph exposes to Claude Code hooks
const (
	EnvRalphActive         = "RALPH_ACTIVE"          // "true" when Ralph is running
	EnvRalphIteration      = "RALPH_ITERATION"       // Current iteration number
	EnvRalphMaxIter        = "RALPH_MAX_ITERATIONS"  // Maximum iterations
	EnvRalphStoryID        = "RALPH_STORY_ID"        // Current story ID being worked on
	EnvRalphStoryTitle     = "RALPH_STORY_TITLE"     // Current story title
	EnvRalphBranch         = "RALPH_BRANCH"          // Git branch from PRD
	EnvRalphPRDPath        = "RALPH_PRD_PATH"        // Path to PRD file
	EnvRalphProgressPath   = "RALPH_PROGRESS_PATH"   // Path to progress file
	EnvRalphPromptPath     = "RALPH_PROMPT_PATH"     // Path to prompt file
	EnvRalphTotalStories   = "RALPH_TOTAL_STORIES"   // Total number of stories
	EnvRalphDoneStories    = "RALPH_DONE_STORIES"    // Number of completed stories
	EnvRalphPendingStories = "RALPH_PENDING_STORIES" // Number of pending stories
	EnvRalphAgentType      = "RALPH_AGENT_TYPE"      // Agent type (claude-code, amp, etc.)
)

// RalphEnv holds Ralph state to expose via environment variables
type RalphEnv struct {
	Active         bool
	Iteration      int
	MaxIterations  int
	StoryID        string
	StoryTitle     string
	Branch         string
	PRDPath        string
	ProgressPath   string
	PromptPath     string
	TotalStories   int
	DoneStories    int
	PendingStories int
	AgentType      string
}

// ToEnvVars converts RalphEnv to a map of environment variables
func (r *RalphEnv) ToEnvVars() map[string]string {
	env := map[string]string{
		EnvRalphActive:         strconv.FormatBool(r.Active),
		EnvRalphIteration:      strconv.Itoa(r.Iteration),
		EnvRalphMaxIter:        strconv.Itoa(r.MaxIterations),
		EnvRalphStoryID:        r.StoryID,
		EnvRalphStoryTitle:     r.StoryTitle,
		EnvRalphBranch:         r.Branch,
		EnvRalphPRDPath:        r.PRDPath,
		EnvRalphProgressPath:   r.ProgressPath,
		EnvRalphPromptPath:     r.PromptPath,
		EnvRalphTotalStories:   strconv.Itoa(r.TotalStories),
		EnvRalphDoneStories:    strconv.Itoa(r.DoneStories),
		EnvRalphPendingStories: strconv.Itoa(r.PendingStories),
		EnvRalphAgentType:      r.AgentType,
	}
	return env
}

// SetEnv sets all Ralph environment variables in the current process
func (r *RalphEnv) SetEnv() error {
	for k, v := range r.ToEnvVars() {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("failed to set %s: %w", k, err)
		}
	}
	return nil
}

// ClearEnv clears all Ralph environment variables
func ClearEnv() {
	vars := []string{
		EnvRalphActive,
		EnvRalphIteration,
		EnvRalphMaxIter,
		EnvRalphStoryID,
		EnvRalphStoryTitle,
		EnvRalphBranch,
		EnvRalphPRDPath,
		EnvRalphProgressPath,
		EnvRalphPromptPath,
		EnvRalphTotalStories,
		EnvRalphDoneStories,
		EnvRalphPendingStories,
		EnvRalphAgentType,
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

// GetRalphEnvFromOS reads Ralph state from current environment variables
// This is useful for hook scripts to access Ralph state
func GetRalphEnvFromOS() *RalphEnv {
	iteration, _ := strconv.Atoi(os.Getenv(EnvRalphIteration))
	maxIter, _ := strconv.Atoi(os.Getenv(EnvRalphMaxIter))
	total, _ := strconv.Atoi(os.Getenv(EnvRalphTotalStories))
	done, _ := strconv.Atoi(os.Getenv(EnvRalphDoneStories))
	pending, _ := strconv.Atoi(os.Getenv(EnvRalphPendingStories))

	return &RalphEnv{
		Active:         os.Getenv(EnvRalphActive) == "true",
		Iteration:      iteration,
		MaxIterations:  maxIter,
		StoryID:        os.Getenv(EnvRalphStoryID),
		StoryTitle:     os.Getenv(EnvRalphStoryTitle),
		Branch:         os.Getenv(EnvRalphBranch),
		PRDPath:        os.Getenv(EnvRalphPRDPath),
		ProgressPath:   os.Getenv(EnvRalphProgressPath),
		PromptPath:     os.Getenv(EnvRalphPromptPath),
		TotalStories:   total,
		DoneStories:    done,
		PendingStories: pending,
		AgentType:      os.Getenv(EnvRalphAgentType),
	}
}

// IsRalphActive checks if Ralph is currently active (from environment)
func IsRalphActive() bool {
	return os.Getenv(EnvRalphActive) == "true"
}
