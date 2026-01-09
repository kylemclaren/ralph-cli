package loop

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/agent"
	"github.com/kylemclaren/ralph/internal/claudecode"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/hooks"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/kylemclaren/ralph/internal/progress"
	"github.com/kylemclaren/ralph/internal/prompt"
)

// Loop manages the Ralph execution loop
type Loop struct {
	Config   *config.Config
	Agent    *agent.Agent
	Hooks    *hooks.Runner
	PRD      *prd.PRD
	Progress *progress.Progress
	Prompt   string

	// State
	Iteration       int
	StartTime       time.Time
	StoriesComplete int
}

// Result holds the result of a loop execution
type Result struct {
	Success         bool
	Iterations      int
	StoriesComplete int
	Duration        time.Duration
	Error           error
	Reason          string // "complete", "max_iterations", "error"
}

// New creates a new loop
func New(cfg *config.Config) (*Loop, error) {
	// Create agent
	cmd, args, err := cfg.GetAgentCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent command: %w", err)
	}

	ag := agent.New(cfg.Agent.Type, cmd, args, cfg.Agent.Timeout)

	// Check agent is available
	if !ag.Available() {
		return nil, fmt.Errorf("agent command '%s' not found in PATH", cmd)
	}

	// Create hooks runner
	hooksRunner := hooks.New(cfg.Hooks.Enabled)
	hooksRunner.SetHooks(
		cfg.Hooks.OnStart,
		cfg.Hooks.OnIteration,
		cfg.Hooks.OnComplete,
		cfg.Hooks.OnFailure,
	)

	return &Loop{
		Config: cfg,
		Agent:  ag,
		Hooks:  hooksRunner,
	}, nil
}

// Load loads the PRD, progress, and prompt files
func (l *Loop) Load() error {
	var err error

	// Load PRD
	l.PRD, err = prd.Load(l.Config.Paths.PRD)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	// Load progress
	l.Progress, err = progress.Load(l.Config.Paths.Progress)
	if err != nil {
		return fmt.Errorf("failed to load progress: %w", err)
	}

	// Load prompt template
	l.Prompt, err = prompt.Load(l.Config.Paths.Prompt)
	if err != nil {
		return fmt.Errorf("failed to load prompt: %w", err)
	}

	return nil
}

// Run executes the Ralph loop
func (l *Loop) Run(ctx context.Context) *Result {
	l.StartTime = time.Now()

	result := &Result{}

	// Check if PRD is already complete
	if l.PRD.IsComplete() {
		result.Success = true
		result.Reason = "complete"
		result.Duration = time.Since(l.StartTime)
		color.Green("All stories already complete!")
		return result
	}

	// Run onStart hooks
	nextStory := l.PRD.NextStory()
	storyID := ""
	if nextStory != nil {
		storyID = nextStory.ID
	}

	if err := l.Hooks.RunOnStart(ctx, 0, storyID); err != nil {
		result.Error = fmt.Errorf("onStart hook failed: %w", err)
		result.Reason = "error"
		return result
	}

	// Main loop
	for l.Iteration = 1; l.Iteration <= l.Config.Loop.MaxIterations; l.Iteration++ {
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			result.Reason = "cancelled"
			result.Iterations = l.Iteration
			return result
		default:
		}

		// Run iteration
		iterResult := l.runIteration(ctx)

		if iterResult.Error != nil {
			result.Error = iterResult.Error
			result.Reason = "error"
			result.Iterations = l.Iteration
			_ = l.Hooks.RunOnFailure(ctx, l.Iteration, iterResult.Error.Error())
			return result
		}

		if iterResult.Complete {
			result.Success = true
			result.Reason = "complete"
			result.Iterations = l.Iteration
			result.StoriesComplete = l.StoriesComplete
			result.Duration = time.Since(l.StartTime)

			_ = l.Hooks.RunOnComplete(ctx, l.Iteration, l.StoriesComplete)

			color.Green("\n✅ All stories complete!")
			fmt.Printf("   Iterations: %d\n", l.Iteration)
			fmt.Printf("   Duration: %v\n", result.Duration.Round(time.Second))
			return result
		}

		// Sleep between iterations
		if l.Iteration < l.Config.Loop.MaxIterations {
			time.Sleep(l.Config.Loop.SleepBetween)
		}
	}

	// Max iterations reached
	result.Reason = "max_iterations"
	result.Iterations = l.Iteration - 1
	result.Duration = time.Since(l.StartTime)
	_ = l.Hooks.RunOnFailure(ctx, l.Iteration, "max iterations reached")

	color.Yellow("\n⚠️  Max iterations reached (%d)", l.Config.Loop.MaxIterations)
	return result
}

// IterationResult holds the result of a single iteration
type IterationResult struct {
	Complete bool
	Error    error
}

// runIteration runs a single loop iteration
func (l *Loop) runIteration(ctx context.Context) *IterationResult {
	result := &IterationResult{}

	// Reload PRD to get latest state
	newPRD, err := prd.Load(l.Config.Paths.PRD)
	if err != nil {
		result.Error = fmt.Errorf("failed to reload PRD: %w", err)
		return result
	}
	l.PRD = newPRD

	// Check if complete
	if l.PRD.IsComplete() {
		result.Complete = true
		return result
	}

	// Get next story
	nextStory := l.PRD.NextStory()
	if nextStory == nil {
		result.Complete = true
		return result
	}

	// Print iteration header
	total, completed, pending := l.PRD.Stats()
	fmt.Println()
	color.Cyan("═══════════════════════════════════════════════════════════════")
	color.Cyan("  Iteration %d/%d | Stories: %d/%d complete | Next: %s",
		l.Iteration, l.Config.Loop.MaxIterations, completed, total, nextStory.ID)
	color.Cyan("═══════════════════════════════════════════════════════════════")
	fmt.Printf("  📋 %s: %s\n", nextStory.ID, nextStory.Title)
	fmt.Println()

	// Run onIteration hooks
	if err := l.Hooks.RunOnIteration(ctx, l.Iteration, nextStory.ID); err != nil {
		result.Error = fmt.Errorf("onIteration hook failed: %w", err)
		return result
	}

	// Reload progress
	l.Progress, err = progress.Load(l.Config.Paths.Progress)
	if err != nil {
		result.Error = fmt.Errorf("failed to reload progress: %w", err)
		return result
	}

	// Build prompt
	templateData, err := prompt.BuildTemplateData(l.PRD, l.Progress)
	if err != nil {
		result.Error = fmt.Errorf("failed to build template data: %w", err)
		return result
	}

	renderedPrompt, err := prompt.Render(l.Prompt, templateData)
	if err != nil {
		result.Error = fmt.Errorf("failed to render prompt: %w", err)
		return result
	}

	// Set Ralph environment variables for the agent
	// This allows Claude Code hooks (and other agents) to access Ralph state
	ralphEnv := &claudecode.RalphEnv{
		Active:         true,
		Iteration:      l.Iteration,
		MaxIterations:  l.Config.Loop.MaxIterations,
		StoryID:        nextStory.ID,
		StoryTitle:     nextStory.Title,
		Branch:         l.PRD.BranchName,
		PRDPath:        l.Config.Paths.PRD,
		ProgressPath:   l.Config.Paths.Progress,
		PromptPath:     l.Config.Paths.Prompt,
		TotalStories:   total,
		DoneStories:    completed,
		PendingStories: pending,
		AgentType:      l.Config.Agent.Type,
	}
	l.Agent.SetEnv(ralphEnv.ToEnvVars())

	// Execute agent
	agentResult, err := l.Agent.Execute(ctx, renderedPrompt)
	if err != nil {
		result.Error = fmt.Errorf("agent execution failed: %w", err)
		return result
	}

	// Check for completion
	if agentResult.IsComplete {
		result.Complete = true
		l.StoriesComplete = completed + pending // All done
	}

	// Update completed count
	newPRD, _ = prd.Load(l.Config.Paths.PRD)
	if newPRD != nil {
		_, newCompleted, _ := newPRD.Stats()
		if newCompleted > completed {
			l.StoriesComplete = newCompleted
		}
	}

	return result
}

// RunOnce runs a single iteration (human-in-the-loop mode)
func (l *Loop) RunOnce(ctx context.Context) *IterationResult {
	l.Iteration = 1
	return l.runIteration(ctx)
}
