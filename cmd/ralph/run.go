package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/loop"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the Ralph loop",
	Long: `Start the Ralph autonomous coding loop.

Ralph will:
1. Read the PRD and find the highest priority pending story
2. Execute your configured AI agent with the prompt
3. Check if all stories are complete
4. Repeat until done or max iterations reached

Examples:
  ralph run                    # Run with default settings
  ralph run --max-iterations 10  # Limit to 10 iterations
  ralph run --once             # Run a single iteration (human-in-the-loop)
  ralph run --dry-run          # Show what would be executed`,
	RunE: runLoop,
}

var (
	runMaxIterations int
	runOnce          bool
	runDryRun        bool
	runVerbose       bool
)

func init() {
	runCmd.Flags().IntVarP(&runMaxIterations, "max-iterations", "n", 0, "Maximum iterations (overrides config)")
	runCmd.Flags().BoolVar(&runOnce, "once", false, "Run a single iteration (human-in-the-loop mode)")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Show what would be executed without running")
	runCmd.Flags().BoolVarP(&runVerbose, "verbose", "v", false, "Verbose output")
	rootCmd.AddCommand(runCmd)
}

func runLoop(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		// Use defaults if no config
		cfg = config.DefaultConfig()
	}

	// Override from flags
	if runMaxIterations > 0 {
		cfg.Loop.MaxIterations = runMaxIterations
	}

	if runOnce {
		cfg.Loop.MaxIterations = 1
	}

	// Check required files exist
	if _, err := os.Stat(cfg.Paths.PRD); os.IsNotExist(err) {
		return fmt.Errorf("PRD not found at %s. Run 'ralph init' first", cfg.Paths.PRD)
	}
	if _, err := os.Stat(cfg.Paths.Prompt); os.IsNotExist(err) {
		return fmt.Errorf("Prompt not found at %s. Run 'ralph init' first", cfg.Paths.Prompt)
	}

	// Create loop
	l, err := loop.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create loop: %w", err)
	}

	// Load files
	if err := l.Load(); err != nil {
		return err
	}

	// Dry run mode
	if runDryRun {
		return dryRun(cfg, l)
	}

	// Print startup info
	printStartup(cfg, l)

	// Set up context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nInterrupted. Cleaning up...")
		cancel()
	}()

	// Run the loop
	var result *loop.Result
	if runOnce {
		color.Cyan("Running single iteration (human-in-the-loop mode)...")
		fmt.Println()
		iterResult := l.RunOnce(ctx)
		result = &loop.Result{
			Success:    iterResult.Complete || iterResult.Error == nil,
			Iterations: 1,
			Error:      iterResult.Error,
		}
		if iterResult.Complete {
			result.Reason = "complete"
		} else if iterResult.Error != nil {
			result.Reason = "error"
		} else {
			result.Reason = "iteration_complete"
		}
	} else {
		result = l.Run(ctx)
	}

	// Print result
	fmt.Println()
	if result.Error != nil {
		color.Red("Error: %v", result.Error)
		return result.Error
	}

	if result.Success {
		if runOnce {
			color.Green("✓ Iteration complete")
			fmt.Println("  Run 'ralph status' to check progress")
			fmt.Println("  Run 'ralph run --once' for another iteration")
		}
	} else if result.Reason == "max_iterations" {
		color.Yellow("Max iterations reached. Run 'ralph run' to continue.")
	}

	return nil
}

func printStartup(cfg *config.Config, l *loop.Loop) {
	total, completed, pending := l.PRD.Stats()

	fmt.Println()
	color.Cyan("🚀 Starting Ralph")
	fmt.Println()
	fmt.Printf("  Agent:      %s\n", cfg.Agent.Type)
	fmt.Printf("  Branch:     %s\n", l.PRD.BranchName)
	fmt.Printf("  Stories:    %d total, %d pending, %d complete\n", total, pending, completed)
	fmt.Printf("  Max Iter:   %d\n", cfg.Loop.MaxIterations)

	if l.Hooks.HasHooks() {
		fmt.Printf("  Hooks:      enabled\n")
	}

	fmt.Println()

	if next := l.PRD.NextStory(); next != nil {
		color.Cyan("  First story: %s - %s", next.ID, next.Title)
	}

	fmt.Println()
}

func dryRun(cfg *config.Config, l *loop.Loop) error {
	total, completed, pending := l.PRD.Stats()

	fmt.Println()
	color.Cyan("Dry Run - Ralph Configuration")
	fmt.Println()

	fmt.Printf("Agent:\n")
	fmt.Printf("  Type:    %s\n", cfg.Agent.Type)
	cmd, args, _ := cfg.GetAgentCommand()
	fmt.Printf("  Command: %s %v\n", cmd, args)
	fmt.Printf("  Timeout: %s\n", cfg.Agent.Timeout)
	fmt.Println()

	fmt.Printf("Loop:\n")
	fmt.Printf("  Max Iterations: %d\n", cfg.Loop.MaxIterations)
	fmt.Printf("  Sleep Between:  %s\n", cfg.Loop.SleepBetween)
	fmt.Println()

	fmt.Printf("Files:\n")
	fmt.Printf("  PRD:      %s\n", cfg.Paths.PRD)
	fmt.Printf("  Progress: %s\n", cfg.Paths.Progress)
	fmt.Printf("  Prompt:   %s\n", cfg.Paths.Prompt)
	fmt.Println()

	fmt.Printf("PRD Status:\n")
	fmt.Printf("  Branch:    %s\n", l.PRD.BranchName)
	fmt.Printf("  Total:     %d stories\n", total)
	fmt.Printf("  Completed: %d stories\n", completed)
	fmt.Printf("  Pending:   %d stories\n", pending)
	fmt.Println()

	if l.Hooks.HasHooks() {
		fmt.Printf("Hooks:\n")
		if len(cfg.Hooks.OnStart) > 0 {
			fmt.Printf("  onStart:     %v\n", cfg.Hooks.OnStart)
		}
		if len(cfg.Hooks.OnIteration) > 0 {
			fmt.Printf("  onIteration: %v\n", cfg.Hooks.OnIteration)
		}
		if len(cfg.Hooks.OnComplete) > 0 {
			fmt.Printf("  onComplete:  %v\n", cfg.Hooks.OnComplete)
		}
		if len(cfg.Hooks.OnFailure) > 0 {
			fmt.Printf("  onFailure:   %v\n", cfg.Hooks.OnFailure)
		}
		fmt.Println()
	}

	if next := l.PRD.NextStory(); next != nil {
		fmt.Printf("Next Story:\n")
		fmt.Printf("  %s\n", next.FormatForDisplay())
	} else {
		color.Green("All stories complete!")
	}

	return nil
}
