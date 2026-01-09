package main

import (
	"fmt"
	"os"

	"github.com/kylemclaren/ralph/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "ralph",
	Short: "Ralph - Autonomous AI coding loop",
	Long: `Ralph is an autonomous AI coding loop that ships features while you sleep.

It runs your AI coding agent (Claude Code, Amp, etc.) repeatedly until all
tasks in your PRD are complete. Each iteration is a fresh context window,
with memory persisting via git history and text files.

Get started:
  ralph init          Initialize Ralph in your project
  ralph add           Add a user story to the PRD
  ralph run           Start the Ralph loop
  ralph status        Check progress`,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./ralph.yaml)")
	rootCmd.PersistentFlags().String("agent", "", "agent type: claude-code, amp, opencode, codex, custom")
	rootCmd.PersistentFlags().Int("max-iterations", 0, "maximum loop iterations")
	rootCmd.PersistentFlags().String("prd", "", "path to PRD file")
	rootCmd.PersistentFlags().String("progress", "", "path to progress file")
	rootCmd.PersistentFlags().String("prompt", "", "path to prompt template")

	// Bind flags to viper
	_ = viper.BindPFlag("agent.type", rootCmd.PersistentFlags().Lookup("agent"))
	_ = viper.BindPFlag("loop.maxIterations", rootCmd.PersistentFlags().Lookup("max-iterations"))
	_ = viper.BindPFlag("paths.prd", rootCmd.PersistentFlags().Lookup("prd"))
	_ = viper.BindPFlag("paths.progress", rootCmd.PersistentFlags().Lookup("progress"))
	_ = viper.BindPFlag("paths.prompt", rootCmd.PersistentFlags().Lookup("prompt"))
}

func initConfig() {
	_, err := config.Load(cfgFile)
	if err != nil {
		// Config file not found is OK for init command
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
