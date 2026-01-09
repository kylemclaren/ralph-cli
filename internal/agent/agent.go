package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Agent represents an AI coding agent
type Agent struct {
	Name    string
	Command string
	Args    []string
	Timeout time.Duration
}

// Result holds the result of an agent execution
type Result struct {
	Output     string
	ExitCode   int
	Duration   time.Duration
	IsComplete bool // true if output contains <promise>COMPLETE</promise>
	Error      error
}

// New creates a new agent
func New(name, command string, args []string, timeout time.Duration) *Agent {
	return &Agent{
		Name:    name,
		Command: command,
		Args:    args,
		Timeout: timeout,
	}
}

// Execute runs the agent with the given prompt
func (a *Agent) Execute(ctx context.Context, prompt string) (*Result, error) {
	start := time.Now()

	// Create context with timeout
	if a.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.Timeout)
		defer cancel()
	}

	// Build the command with prompt
	args := a.buildArgs(prompt)
	cmd := exec.CommandContext(ctx, a.Command, args...)

	// Capture output while also streaming to stdout/stderr
	var outputBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &outputBuf)
	cmd.Stdin = strings.NewReader(prompt)

	// Run the command
	err := cmd.Run()

	result := &Result{
		Output:   outputBuf.String(),
		Duration: time.Since(start),
	}

	// Check exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("agent timed out after %v", a.Timeout)
			result.ExitCode = -1
		} else {
			result.Error = err
			result.ExitCode = -1
		}
	}

	// Check for completion marker
	result.IsComplete = strings.Contains(result.Output, "<promise>COMPLETE</promise>")

	return result, nil
}

// ExecuteWithStdin runs the agent by piping prompt via stdin
func (a *Agent) ExecuteWithStdin(ctx context.Context, prompt string) (*Result, error) {
	start := time.Now()

	if a.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.Timeout)
		defer cancel()
	}

	// For agents that accept prompt via stdin (like amp with piped input)
	cmd := exec.CommandContext(ctx, a.Command, a.Args...)

	var outputBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &outputBuf)
	cmd.Stdin = strings.NewReader(prompt)

	err := cmd.Run()

	result := &Result{
		Output:   outputBuf.String(),
		Duration: time.Since(start),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("agent timed out after %v", a.Timeout)
			result.ExitCode = -1
		} else {
			result.Error = err
			result.ExitCode = -1
		}
	}

	result.IsComplete = strings.Contains(result.Output, "<promise>COMPLETE</promise>")

	return result, nil
}

// buildArgs builds command arguments, adding prompt flag if needed
func (a *Agent) buildArgs(prompt string) []string {
	args := make([]string, len(a.Args))
	copy(args, a.Args)

	// Add prompt as -p flag for most agents
	switch a.Name {
	case "claude-code":
		args = append(args, "-p", prompt)
	case "amp":
		// amp uses stdin, handled separately
		args = append(args, "-p", prompt)
	case "opencode":
		args = append(args, "-p", prompt)
	case "codex":
		args = append(args, "-p", prompt)
	default:
		// For custom agents, assume -p flag
		args = append(args, "-p", prompt)
	}

	return args
}

// CommandString returns the full command string for display
func (a *Agent) CommandString() string {
	return fmt.Sprintf("%s %s", a.Command, strings.Join(a.Args, " "))
}

// Available checks if the agent command is available
func (a *Agent) Available() bool {
	_, err := exec.LookPath(a.Command)
	return err == nil
}

// AgentType returns the agent type from a string
func AgentType(t string) string {
	switch strings.ToLower(t) {
	case "claude", "claude-code", "claudecode":
		return "claude-code"
	case "amp":
		return "amp"
	case "opencode", "open-code":
		return "opencode"
	case "codex":
		return "codex"
	default:
		return t
	}
}
