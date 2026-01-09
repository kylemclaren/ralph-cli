package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HookType represents the type of hook
type HookType string

const (
	HookOnStart     HookType = "onStart"
	HookOnIteration HookType = "onIteration"
	HookOnComplete  HookType = "onComplete"
	HookOnFailure   HookType = "onFailure"
)

// Runner executes hooks
type Runner struct {
	OnStart     []string
	OnIteration []string
	OnComplete  []string
	OnFailure   []string
	Enabled     bool
	Verbose     bool
}

// New creates a new hook runner
func New(enabled bool) *Runner {
	return &Runner{
		Enabled: enabled,
	}
}

// SetHooks sets hooks from config
func (r *Runner) SetHooks(onStart, onIteration, onComplete, onFailure []string) {
	r.OnStart = onStart
	r.OnIteration = onIteration
	r.OnComplete = onComplete
	r.OnFailure = onFailure
}

// Run executes hooks of the given type
func (r *Runner) Run(ctx context.Context, hookType HookType, env map[string]string) error {
	if !r.Enabled {
		return nil
	}

	var hooks []string
	switch hookType {
	case HookOnStart:
		hooks = r.OnStart
	case HookOnIteration:
		hooks = r.OnIteration
	case HookOnComplete:
		hooks = r.OnComplete
	case HookOnFailure:
		hooks = r.OnFailure
	default:
		return fmt.Errorf("unknown hook type: %s", hookType)
	}

	for _, hook := range hooks {
		if err := r.runSingle(ctx, hook, env); err != nil {
			return fmt.Errorf("hook %s failed: %w", hook, err)
		}
	}

	return nil
}

// runSingle executes a single hook command
func (r *Runner) runSingle(ctx context.Context, command string, env map[string]string) error {
	if command == "" {
		return nil
	}

	if r.Verbose {
		fmt.Printf("  Running hook: %s\n", command)
	}

	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	return cmd.Run()
}

// RunOnStart runs onStart hooks
func (r *Runner) RunOnStart(ctx context.Context, iteration int, storyID string) error {
	return r.Run(ctx, HookOnStart, map[string]string{
		"RALPH_ITERATION": fmt.Sprintf("%d", iteration),
		"RALPH_STORY_ID":  storyID,
		"RALPH_HOOK":      string(HookOnStart),
	})
}

// RunOnIteration runs onIteration hooks
func (r *Runner) RunOnIteration(ctx context.Context, iteration int, storyID string) error {
	return r.Run(ctx, HookOnIteration, map[string]string{
		"RALPH_ITERATION": fmt.Sprintf("%d", iteration),
		"RALPH_STORY_ID":  storyID,
		"RALPH_HOOK":      string(HookOnIteration),
	})
}

// RunOnComplete runs onComplete hooks
func (r *Runner) RunOnComplete(ctx context.Context, iterations int, storiesCompleted int) error {
	return r.Run(ctx, HookOnComplete, map[string]string{
		"RALPH_ITERATIONS":        fmt.Sprintf("%d", iterations),
		"RALPH_STORIES_COMPLETED": fmt.Sprintf("%d", storiesCompleted),
		"RALPH_HOOK":              string(HookOnComplete),
	})
}

// RunOnFailure runs onFailure hooks
func (r *Runner) RunOnFailure(ctx context.Context, iteration int, reason string) error {
	return r.Run(ctx, HookOnFailure, map[string]string{
		"RALPH_ITERATION":      fmt.Sprintf("%d", iteration),
		"RALPH_FAILURE_REASON": reason,
		"RALPH_HOOK":           string(HookOnFailure),
	})
}

// HasHooks returns true if any hooks are configured
func (r *Runner) HasHooks() bool {
	return len(r.OnStart) > 0 ||
		len(r.OnIteration) > 0 ||
		len(r.OnComplete) > 0 ||
		len(r.OnFailure) > 0
}
