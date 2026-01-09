package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show PRD status and progress",
	Long:  `Display the current status of all user stories in the PRD.`,
	RunE:  runStatus,
}

var (
	statusJSON    bool
	statusPending bool
	statusDone    bool
)

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
	statusCmd.Flags().BoolVar(&statusPending, "pending", false, "Show only pending stories")
	statusCmd.Flags().BoolVar(&statusDone, "done", false, "Show only completed stories")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Check if PRD exists
	if _, err := os.Stat(cfg.Paths.PRD); os.IsNotExist(err) {
		color.Yellow("No PRD found. Run 'ralph init' to get started.")
		return nil
	}

	// Load PRD
	p, err := prd.Load(cfg.Paths.PRD)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	// JSON output
	if statusJSON {
		json, err := p.ToJSON()
		if err != nil {
			return err
		}
		fmt.Println(json)
		return nil
	}

	// Get stats
	total, completed, pending := p.Stats()

	// Print header
	fmt.Println()
	color.Cyan("═══════════════════════════════════════════════════════════════")
	color.Cyan("  Ralph Status")
	color.Cyan("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Print stats
	fmt.Printf("  Branch: %s\n", p.BranchName)
	fmt.Printf("  Total:  %d stories\n", total)

	if completed > 0 {
		color.Green("  Done:   %d stories", completed)
	} else {
		fmt.Printf("  Done:   %d stories\n", completed)
	}

	if pending > 0 {
		color.Yellow("  Pending: %d stories", pending)
	} else {
		fmt.Printf("  Pending: %d stories\n", pending)
	}

	// Progress bar
	if total > 0 {
		fmt.Println()
		printProgressBar(completed, total)
	}

	fmt.Println()

	// Filter stories
	var stories []prd.UserStory
	if statusPending {
		stories = p.PendingStories()
		if len(stories) == 0 {
			color.Green("  All stories complete! 🎉")
			return nil
		}
	} else if statusDone {
		stories = p.CompletedStories()
		if len(stories) == 0 {
			color.Yellow("  No completed stories yet.")
			return nil
		}
	} else {
		stories = p.UserStories
	}

	// Print stories
	if len(stories) == 0 {
		color.Yellow("  No stories in PRD. Run 'ralph add' to add stories.")
		return nil
	}

	fmt.Println("  Stories:")
	fmt.Println("  " + strings.Repeat("─", 60))

	for _, story := range stories {
		printStory(story)
	}

	// Next story hint
	if next := p.NextStory(); next != nil && !statusDone {
		fmt.Println()
		color.Cyan("  Next up: %s - %s", next.ID, next.Title)
	}

	fmt.Println()
	return nil
}

func printStory(s prd.UserStory) {
	// Status icon
	var status string
	if s.Passes {
		status = color.GreenString("✓")
	} else {
		status = color.YellowString("○")
	}

	// Priority badge
	priority := fmt.Sprintf("P%d", s.Priority)

	// Print story line
	fmt.Printf("  %s [%s] %s: %s\n", status, priority, s.ID, s.Title)

	// Print acceptance criteria if pending
	if !s.Passes && len(s.AcceptanceCriteria) > 0 {
		for _, ac := range s.AcceptanceCriteria {
			fmt.Printf("      • %s\n", ac)
		}
	}
}

func printProgressBar(completed, total int) {
	width := 40
	filled := (completed * width) / total
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	percentage := (completed * 100) / total

	var coloredBar string
	if percentage == 100 {
		coloredBar = color.GreenString(bar)
	} else if percentage >= 50 {
		coloredBar = color.YellowString(bar)
	} else {
		coloredBar = color.RedString(bar)
	}

	fmt.Printf("  [%s] %d%%\n", coloredBar, percentage)
}
