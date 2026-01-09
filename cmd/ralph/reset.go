package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset <story-id>",
	Short: "Mark a story as pending (not passing)",
	Long: `Mark a user story as pending/not passing.

Examples:
  ralph reset US-001
  ralph reset us-001    # Case insensitive
  ralph reset --all     # Reset all stories`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReset,
}

var resetAll bool

func init() {
	resetCmd.Flags().BoolVar(&resetAll, "all", false, "Reset all stories to pending")
	rootCmd.AddCommand(resetCmd)
}

func runReset(cmd *cobra.Command, args []string) error {
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

	if resetAll {
		// Reset all stories
		count := 0
		for i := range p.UserStories {
			if p.UserStories[i].Passes {
				p.UserStories[i].Passes = false
				count++
			}
		}

		if err := p.Save(cfg.Paths.PRD); err != nil {
			return fmt.Errorf("failed to save PRD: %w", err)
		}

		if count == 0 {
			color.Yellow("No completed stories to reset")
		} else {
			color.Green("✓ Reset %d stories to pending", count)
		}
		return nil
	}

	// Reset single story
	if len(args) == 0 {
		return fmt.Errorf("story ID required (or use --all)")
	}

	storyID := args[0]

	// Find and update story
	story := p.GetStory(storyID)
	if story == nil {
		return fmt.Errorf("story %s not found", storyID)
	}

	if !story.Passes {
		color.Yellow("Story %s is already pending", story.ID)
		return nil
	}

	if err := p.MarkPending(storyID); err != nil {
		return err
	}

	// Save PRD
	if err := p.Save(cfg.Paths.PRD); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	color.Green("✓ Reset %s to pending: %s", story.ID, story.Title)

	return nil
}
