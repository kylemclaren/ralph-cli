package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <story-id>",
	Short: "Mark a story as complete (passing)",
	Long: `Mark a user story as complete/passing.

Examples:
  ralph done US-001
  ralph done us-001    # Case insensitive`,
	Args: cobra.ExactArgs(1),
	RunE: runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(cmd *cobra.Command, args []string) error {
	storyID := args[0]

	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Load PRD
	p, err := prd.Load(cfg.Paths.PRD)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	// Find and update story
	story := p.GetStory(storyID)
	if story == nil {
		return fmt.Errorf("story %s not found", storyID)
	}

	if story.Passes {
		color.Yellow("Story %s is already marked as done", story.ID)
		return nil
	}

	if err := p.MarkDone(storyID); err != nil {
		return err
	}

	// Save PRD
	if err := p.Save(cfg.Paths.PRD); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	color.Green("✓ Marked %s as done: %s", story.ID, story.Title)

	// Show progress
	total, completed, _ := p.Stats()
	fmt.Printf("  Progress: %d/%d complete\n", completed, total)

	if p.IsComplete() {
		color.Green("\n🎉 All stories complete!")
	}

	return nil
}
