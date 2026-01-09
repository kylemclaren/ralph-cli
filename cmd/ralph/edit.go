package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <story-id>",
	Short: "Edit an existing user story",
	Long: `Edit an existing user story in the PRD.

Examples:
  ralph edit US-001                    # Interactive edit
  ralph edit US-001 -t "New title"     # Update title only
  ralph edit US-001 -p 1               # Update priority only`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var (
	editTitle       string
	editDescription string
	editPriority    int
	editNotes       string
)

func init() {
	editCmd.Flags().StringVarP(&editTitle, "title", "t", "", "New title")
	editCmd.Flags().StringVarP(&editDescription, "description", "d", "", "New description")
	editCmd.Flags().IntVarP(&editPriority, "priority", "p", 0, "New priority")
	editCmd.Flags().StringVarP(&editNotes, "notes", "n", "", "New notes")
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
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

	// Check if any flags were provided
	flagsProvided := editTitle != "" || editDescription != "" || editPriority != 0 || editNotes != ""

	if flagsProvided {
		// Update from flags
		if editTitle != "" {
			story.Title = editTitle
		}
		if editDescription != "" {
			story.Description = editDescription
		}
		if editPriority != 0 {
			story.Priority = editPriority
		}
		if editNotes != "" {
			story.Notes = editNotes
		}
	} else {
		// Interactive edit
		if err := interactiveEdit(story); err != nil {
			return err
		}
	}

	// Save PRD
	if err := p.Save(cfg.Paths.PRD); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	color.Green("✓ Updated story: %s", story.ID)
	fmt.Println(story.FormatForDisplay())

	return nil
}

func interactiveEdit(story *prd.UserStory) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	color.Cyan("Edit Story: %s", story.ID)
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("(Press Enter to keep current value)")
	fmt.Println()

	// Title
	fmt.Printf("Title [%s]: ", story.Title)
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)
	if title != "" {
		story.Title = title
	}

	// Description
	fmt.Printf("Description [%s]: ", story.Description)
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)
	if desc != "" {
		story.Description = desc
	}

	// Priority
	fmt.Printf("Priority [%d]: ", story.Priority)
	priorityStr, _ := reader.ReadString('\n')
	priorityStr = strings.TrimSpace(priorityStr)
	if priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			story.Priority = priority
		}
	}

	// Notes
	fmt.Printf("Notes [%s]: ", story.Notes)
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)
	if notes != "" {
		story.Notes = notes
	}

	return nil
}
