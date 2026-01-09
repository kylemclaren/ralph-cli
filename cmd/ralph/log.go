package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/progress"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "View or append to the progress log",
	Long: `View the progress log or append entries.

Examples:
  ralph log                    # View the progress log
  ralph log --edit             # Edit the progress log
  ralph log --append "Note"    # Append a note
  ralph log --patterns         # Show codebase patterns section
  ralph log --clear            # Clear the progress log`,
	RunE: runLog,
}

var (
	logEdit     bool
	logAppend   string
	logPatterns bool
	logClear    bool
	logTail     int
)

func init() {
	logCmd.Flags().BoolVarP(&logEdit, "edit", "e", false, "Edit the progress log")
	logCmd.Flags().StringVarP(&logAppend, "append", "a", "", "Append a note to the log")
	logCmd.Flags().BoolVarP(&logPatterns, "patterns", "p", false, "Show codebase patterns section")
	logCmd.Flags().BoolVar(&logClear, "clear", false, "Clear and reset the progress log")
	logCmd.Flags().IntVarP(&logTail, "tail", "t", 0, "Show last N lines")
	rootCmd.AddCommand(logCmd)
}

func runLog(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	progressPath := cfg.Paths.Progress

	// Clear the log
	if logClear {
		if _, err := progress.Create(progressPath); err != nil {
			return fmt.Errorf("failed to clear progress log: %w", err)
		}
		color.Green("✓ Progress log cleared")
		return nil
	}

	// Edit mode
	if logEdit {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		cmd := exec.Command(editor, progressPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	// Check if progress file exists
	if !progress.Exists(progressPath) {
		color.Yellow("Progress log not found at %s", progressPath)
		fmt.Println("Run 'ralph init' to create one")
		return nil
	}

	// Load progress
	prog, err := progress.Load(progressPath)
	if err != nil {
		return fmt.Errorf("failed to load progress: %w", err)
	}

	// Append mode
	if logAppend != "" {
		prog.Append(fmt.Sprintf("\n**Note:** %s\n", logAppend))
		if err := prog.Save(); err != nil {
			return fmt.Errorf("failed to save progress: %w", err)
		}
		color.Green("✓ Appended note to progress log")
		return nil
	}

	// Patterns mode
	if logPatterns {
		patterns := prog.GetCodebasePatterns()
		if patterns == "" {
			color.Yellow("No codebase patterns found in progress log")
			return nil
		}
		fmt.Println("## Codebase Patterns")
		fmt.Println(patterns)
		return nil
	}

	// Default: view the log
	content := prog.Content

	// Tail mode
	if logTail > 0 {
		lines := strings.Split(content, "\n")
		if len(lines) > logTail {
			lines = lines[len(lines)-logTail:]
		}
		content = strings.Join(lines, "\n")
	}

	fmt.Println(content)
	return nil
}

// interactiveAppend prompts for multiline input
func interactiveAppend() (string, error) {
	fmt.Println("Enter your note (Ctrl+D to finish):")
	fmt.Println(strings.Repeat("─", 40))

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}
