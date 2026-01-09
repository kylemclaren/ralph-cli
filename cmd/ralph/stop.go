package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/kylemclaren/ralph/internal/pidfile"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running Ralph loop",
	Long: `Stop a running Ralph loop gracefully.

Sends SIGTERM to the running Ralph process, allowing it to finish the
current iteration cleanly before exiting.

Use --force to send SIGKILL for immediate termination.

Examples:
  ralph stop          # Graceful stop (SIGTERM)
  ralph stop --force  # Force stop (SIGKILL)`,
	RunE: stopLoop,
}

var (
	stopForce bool
)

func init() {
	stopCmd.Flags().BoolVarP(&stopForce, "force", "f", false, "Force stop (SIGKILL)")
	rootCmd.AddCommand(stopCmd)
}

func stopLoop(cmd *cobra.Command, args []string) error {
	pf := pidfile.New("")

	running, pid := pf.IsRunning()
	if !running {
		color.Yellow("Ralph is not running")
		return nil
	}

	fmt.Printf("Found Ralph process (PID %d)\n", pid)

	var err error
	if stopForce {
		fmt.Println("Sending SIGKILL...")
		err = pf.Kill()
	} else {
		fmt.Println("Sending SIGTERM for graceful shutdown...")
		err = pf.Stop()
	}

	if err != nil {
		return fmt.Errorf("failed to stop Ralph: %w", err)
	}

	// Wait for process to exit
	fmt.Print("Waiting for process to exit")
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if running, _ := pf.IsRunning(); !running {
			fmt.Println()
			color.Green("Ralph stopped successfully")
			_ = pf.Remove() // Clean up PID file
			return nil
		}
		fmt.Print(".")
	}

	fmt.Println()
	if !stopForce {
		color.Yellow("Process still running. Use --force to kill immediately.")
	} else {
		color.Red("Failed to stop process")
	}

	return nil
}
