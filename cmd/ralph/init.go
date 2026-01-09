package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/kylemclaren/ralph/internal/progress"
	"github.com/kylemclaren/ralph/internal/prompt"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Ralph in the current project",
	Long: `Initialize Ralph by creating the necessary files and directories:
  - ralph.yaml (configuration)
  - .ralph/prd.json (product requirements document)
  - .ralph/progress.txt (progress log)
  - .ralph/prompt.md (agent prompt template)`,
	RunE: runInit,
}

var (
	initBranch  string
	initForce   bool
	initMinimal bool
)

func init() {
	initCmd.Flags().StringVarP(&initBranch, "branch", "b", "ralph/feature", "Git branch name for the feature")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing files")
	initCmd.Flags().BoolVarP(&initMinimal, "minimal", "m", false, "Create minimal config without examples")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Initializing Ralph...")
	fmt.Println()

	// Get default config
	cfg := config.DefaultConfig()

	// Create directories
	if err := cfg.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Track what was created
	var created []string
	var skipped []string

	// Create ralph.yaml
	configPath := "ralph.yaml"
	if !fileExists(configPath) || initForce {
		if err := writeConfigFile(configPath); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		created = append(created, configPath)
	} else {
		skipped = append(skipped, configPath)
	}

	// Create PRD
	if !fileExists(cfg.Paths.PRD) || initForce {
		p := prd.NewPRD(initBranch)
		if !initMinimal {
			// Add example story
			p.AddStory(prd.UserStory{
				Title:       "Example user story",
				Description: "Replace this with your actual user story",
				AcceptanceCriteria: []string{
					"Define clear acceptance criteria",
					"Include testable conditions",
					"typecheck passes",
					"tests pass",
				},
				Priority: 1,
				Passes:   false,
			})
		}
		if err := p.Save(cfg.Paths.PRD); err != nil {
			return fmt.Errorf("failed to create PRD: %w", err)
		}
		created = append(created, cfg.Paths.PRD)
	} else {
		skipped = append(skipped, cfg.Paths.PRD)
	}

	// Create progress.txt
	if !fileExists(cfg.Paths.Progress) || initForce {
		if _, err := progress.Create(cfg.Paths.Progress); err != nil {
			return fmt.Errorf("failed to create progress file: %w", err)
		}
		created = append(created, cfg.Paths.Progress)
	} else {
		skipped = append(skipped, cfg.Paths.Progress)
	}

	// Create prompt.md
	if !fileExists(cfg.Paths.Prompt) || initForce {
		if err := prompt.Create(cfg.Paths.Prompt); err != nil {
			return fmt.Errorf("failed to create prompt file: %w", err)
		}
		created = append(created, cfg.Paths.Prompt)
	} else {
		skipped = append(skipped, cfg.Paths.Prompt)
	}

	// Print results
	if len(created) > 0 {
		color.Green("✓ Created:")
		for _, f := range created {
			fmt.Printf("  - %s\n", f)
		}
	}

	if len(skipped) > 0 {
		color.Yellow("⊘ Skipped (already exists):")
		for _, f := range skipped {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println()
		fmt.Println("  Use --force to overwrite existing files")
	}

	fmt.Println()
	color.Cyan("Next steps:")
	fmt.Println("  1. Edit .ralph/prd.json to add your user stories")
	fmt.Println("  2. Customize .ralph/prompt.md if needed")
	fmt.Println("  3. Run 'ralph status' to see your stories")
	fmt.Println("  4. Run 'ralph run' to start the loop")

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writeConfigFile(path string) error {
	content := `# Ralph Configuration
# See https://github.com/kylemclaren/ralph for documentation

# Agent configuration
agent:
  # Agent type: claude-code, amp, opencode, codex, custom
  type: claude-code
  # Custom command (only if type: custom)
  # command: "my-agent --flag"
  # Additional flags to pass to the agent
  flags: []
  # Maximum time per iteration
  timeout: 30m

# Loop configuration
loop:
  # Maximum number of iterations
  maxIterations: 25
  # Time to sleep between iterations
  sleepBetween: 2s
  # Stop on first failure (default: continue)
  stopOnFirstFailure: false

# File paths (relative to project root)
paths:
  prd: .ralph/prd.json
  progress: .ralph/progress.txt
  prompt: .ralph/prompt.md

# Lifecycle hooks
hooks:
  enabled: true
  # Commands to run before the loop starts
  onStart: []
  # Commands to run before each iteration
  onIteration: []
  # Commands to run when all stories complete
  onComplete: []
  # Commands to run on failure
  onFailure: []

# Notifications (optional)
notifications:
  enabled: false
  # Webhook URL for Slack/Discord notifications
  webhook: ""
`
	return os.WriteFile(path, []byte(content), 0644)
}
