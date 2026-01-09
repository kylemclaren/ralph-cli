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

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user story to the PRD",
	Long: `Add a new user story to the PRD interactively or via flags.

Examples:
  ralph add                                    # Interactive mode
  ralph add -t "Add login form" -p 1           # Quick add with title and priority
  ralph add -t "Feature" -a "Criterion 1" -a "Criterion 2"  # With acceptance criteria`,
	RunE: runAdd,
}

var (
	addTitle              string
	addDescription        string
	addPriority           int
	addAcceptanceCriteria []string
	addInteractive        bool
)

func init() {
	addCmd.Flags().StringVarP(&addTitle, "title", "t", "", "Story title")
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "Story description")
	addCmd.Flags().IntVarP(&addPriority, "priority", "p", 0, "Priority (lower = higher priority)")
	addCmd.Flags().StringArrayVarP(&addAcceptanceCriteria, "acceptance", "a", nil, "Acceptance criteria (can be repeated)")
	addCmd.Flags().BoolVarP(&addInteractive, "interactive", "i", false, "Force interactive mode")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Check if PRD exists
	if _, err := os.Stat(cfg.Paths.PRD); os.IsNotExist(err) {
		return fmt.Errorf("PRD not found. Run 'ralph init' first")
	}

	// Load PRD
	p, err := prd.Load(cfg.Paths.PRD)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	var story prd.UserStory

	// Determine if we should use interactive mode
	useInteractive := addInteractive || (addTitle == "" && len(addAcceptanceCriteria) == 0)

	if useInteractive {
		story, err = interactiveAdd(p)
		if err != nil {
			return err
		}
	} else {
		// Use flags
		if addTitle == "" {
			return fmt.Errorf("title is required (use -t or --title)")
		}

		story = prd.UserStory{
			Title:              addTitle,
			Description:        addDescription,
			AcceptanceCriteria: addAcceptanceCriteria,
			Priority:           addPriority,
			Passes:             false,
		}

		// Set default priority if not specified
		if story.Priority == 0 {
			story.Priority = len(p.UserStories) + 1
		}

		// Add default acceptance criteria if none provided
		if len(story.AcceptanceCriteria) == 0 {
			story.AcceptanceCriteria = []string{
				"typecheck passes",
				"tests pass",
			}
		}
	}

	// Add story to PRD
	p.AddStory(story)

	// Save PRD
	if err := p.Save(cfg.Paths.PRD); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	// Get the story back to show the generated ID
	addedStory := p.UserStories[len(p.UserStories)-1]

	color.Green("✓ Added story: %s", addedStory.ID)
	fmt.Printf("  Title: %s\n", addedStory.Title)
	fmt.Printf("  Priority: %d\n", addedStory.Priority)
	fmt.Printf("  Acceptance Criteria: %d items\n", len(addedStory.AcceptanceCriteria))

	return nil
}

func interactiveAdd(p *prd.PRD) (prd.UserStory, error) {
	reader := bufio.NewReader(os.Stdin)
	story := prd.UserStory{}

	fmt.Println()
	color.Cyan("Add New User Story")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println()

	// Title (required)
	fmt.Print("Title: ")
	title, err := reader.ReadString('\n')
	if err != nil {
		return story, err
	}
	story.Title = strings.TrimSpace(title)
	if story.Title == "" {
		return story, fmt.Errorf("title is required")
	}

	// Description (optional)
	fmt.Print("Description (optional): ")
	desc, err := reader.ReadString('\n')
	if err != nil {
		return story, err
	}
	story.Description = strings.TrimSpace(desc)

	// Priority
	defaultPriority := len(p.UserStories) + 1
	fmt.Printf("Priority [%d]: ", defaultPriority)
	priorityStr, err := reader.ReadString('\n')
	if err != nil {
		return story, err
	}
	priorityStr = strings.TrimSpace(priorityStr)
	if priorityStr == "" {
		story.Priority = defaultPriority
	} else {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return story, fmt.Errorf("invalid priority: %s", priorityStr)
		}
		story.Priority = priority
	}

	// Acceptance criteria
	fmt.Println()
	fmt.Println("Acceptance Criteria (enter each criterion, empty line to finish):")
	fmt.Println("  Tip: Include 'typecheck passes' and 'tests pass' for better feedback loops")
	fmt.Println()

	for i := 1; ; i++ {
		fmt.Printf("  %d. ", i)
		criterion, err := reader.ReadString('\n')
		if err != nil {
			return story, err
		}
		criterion = strings.TrimSpace(criterion)
		if criterion == "" {
			break
		}
		story.AcceptanceCriteria = append(story.AcceptanceCriteria, criterion)
	}

	// Add default criteria if none provided
	if len(story.AcceptanceCriteria) == 0 {
		story.AcceptanceCriteria = []string{
			"typecheck passes",
			"tests pass",
		}
		fmt.Println("  (Added default criteria: typecheck passes, tests pass)")
	}

	story.Passes = false

	return story, nil
}
