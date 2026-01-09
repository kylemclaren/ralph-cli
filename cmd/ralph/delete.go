package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <story-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a user story from the PRD",
	Long: `Delete a user story from the PRD.

Examples:
  ralph delete US-001
  ralph delete US-001 --force   # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var deleteForce bool

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	// Find story
	story := p.GetStory(storyID)
	if story == nil {
		return fmt.Errorf("story %s not found", storyID)
	}

	// Confirm deletion
	if !deleteForce {
		fmt.Printf("Delete story %s: %s?\n", story.ID, story.Title)
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete story
	if err := p.DeleteStory(storyID); err != nil {
		return err
	}

	// Save PRD
	if err := p.Save(cfg.Paths.PRD); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	color.Green("✓ Deleted story: %s", story.ID)

	return nil
}
