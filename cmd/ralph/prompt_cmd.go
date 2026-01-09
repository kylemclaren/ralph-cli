package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/config"
	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/kylemclaren/ralph/internal/progress"
	"github.com/kylemclaren/ralph/internal/prompt"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "View or edit the prompt template",
	Long: `View the current prompt template or edit it.

Examples:
  ralph prompt              # View the prompt template
  ralph prompt --edit       # Edit the prompt template
  ralph prompt --render     # Render the prompt with current PRD/progress
  ralph prompt --reset      # Reset to default prompt`,
	RunE: runPrompt,
}

var (
	promptEdit   bool
	promptRender bool
	promptReset  bool
)

func init() {
	promptCmd.Flags().BoolVarP(&promptEdit, "edit", "e", false, "Edit the prompt template")
	promptCmd.Flags().BoolVarP(&promptRender, "render", "r", false, "Render the prompt with current data")
	promptCmd.Flags().BoolVar(&promptReset, "reset", false, "Reset to default prompt template")
	rootCmd.AddCommand(promptCmd)
}

func runPrompt(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	promptPath := cfg.Paths.Prompt

	// Reset to default
	if promptReset {
		if err := prompt.Create(promptPath); err != nil {
			return fmt.Errorf("failed to reset prompt: %w", err)
		}
		color.Green("✓ Reset prompt template to default")
		return nil
	}

	// Edit mode
	if promptEdit {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		cmd := exec.Command(editor, promptPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	// Check if prompt exists
	if !prompt.Exists(promptPath) {
		color.Yellow("Prompt template not found at %s", promptPath)
		fmt.Println("Run 'ralph init' to create one, or 'ralph prompt --reset' to create default")
		return nil
	}

	// Render mode
	if promptRender {
		return renderPrompt(cfg)
	}

	// Default: view the template
	content, err := prompt.Load(promptPath)
	if err != nil {
		return fmt.Errorf("failed to load prompt: %w", err)
	}

	fmt.Println(content)
	return nil
}

func renderPrompt(cfg *config.Config) error {
	// Load prompt template
	templateContent, err := prompt.Load(cfg.Paths.Prompt)
	if err != nil {
		return fmt.Errorf("failed to load prompt: %w", err)
	}

	// Load PRD
	p, err := prd.Load(cfg.Paths.PRD)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	// Load progress
	prog, err := progress.Load(cfg.Paths.Progress)
	if err != nil {
		return fmt.Errorf("failed to load progress: %w", err)
	}

	// Build template data
	data, err := prompt.BuildTemplateData(p, prog)
	if err != nil {
		return fmt.Errorf("failed to build template data: %w", err)
	}

	// Render
	rendered, err := prompt.Render(templateContent, data)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	fmt.Println(rendered)
	return nil
}
