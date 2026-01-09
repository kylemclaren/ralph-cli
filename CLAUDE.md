# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
mise run build      # Build ralph binary
mise run test       # Run all tests
mise run lint       # Run golangci-lint
mise run install    # Install to GOPATH/bin

# Run a single test
go test -v -run TestName ./internal/prd/

# Run tests for a specific package
go test -v ./internal/agent/
```

## Architecture

Ralph is an autonomous AI coding loop that repeatedly runs AI coding agents (Claude Code, Amp, OpenCode, Codex) until all tasks in a PRD are complete. Each iteration is a fresh context window with memory persisting via git and text files.

### Core Loop (`internal/loop/loop.go`)
The main execution loop that:
1. Loads PRD, progress log, and prompt template
2. Finds highest priority pending story
3. Renders the prompt with current state
4. Executes the configured AI agent
5. Checks for completion marker (`<promise>COMPLETE</promise>`)
6. Repeats until all stories pass or max iterations reached

### Key Packages

- **`internal/agent`** - Agent abstraction for executing AI tools. Handles command building, timeout, environment variables, and output capture. Agents receive prompts via `-p` flag.

- **`internal/prd`** - PRD (Product Requirements Document) management. UserStories have ID, title, acceptance criteria, priority, and passes boolean. NextStory() returns highest priority pending story.

- **`internal/config`** - Configuration via Viper. Supports ralph.yaml file, RALPH_ env vars, and CLI flags. Config precedence: flags > env > file.

- **`internal/hooks`** - Lifecycle hooks (onStart, onIteration, onComplete, onFailure) that run shell commands at key points.

- **`internal/prompt`** - Go text/template rendering for agent prompts. TemplateData includes PRD JSON, progress content, and story counts.

- **`internal/claudecode`** - Environment variable bridge. Sets RALPH_* env vars so Claude Code hooks can access Ralph state (iteration, story ID, etc.).

### CLI Commands (`cmd/ralph/`)
Each file is a Cobra command: init, add, edit, done, reset, delete, status, run, prompt, log, version.

### Data Files (`.ralph/` directory)
- `prd.json` - User stories with acceptance criteria
- `progress.txt` - Learnings log that compounds across iterations
- `prompt.md` - Customizable agent prompt template

### Agent Execution Flow
1. Config determines agent command (e.g., `claude --dangerously-skip-permissions`)
2. Loop builds TemplateData from PRD and progress
3. Prompt template is rendered with current state
4. Agent.Execute() runs command with `-p` flag containing rendered prompt
5. Output is checked for `<promise>COMPLETE</promise>` to signal all stories done
