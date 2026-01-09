# Ralph CLI

An autonomous AI coding loop that ships features while you sleep.

Ralph runs your AI coding agent (Claude Code, Amp, OpenCode, Codex) repeatedly until all tasks in your PRD are complete. Each iteration is a fresh context window, with memory persisting via git history and text files.

Based on the [Ralph methodology](https://ghuntley.com/ralph/) by Geoffrey Huntley.

## Installation

```bash
# Using go install
go install github.com/kylemclaren/ralph-cli/cmd/ralph@latest

# Or clone and build
git clone https://github.com/kylemclaren/ralph-cli.git
cd ralph-cli
make build
```

## Quick Start

```bash
# Initialize Ralph in your project
ralph init

# Add user stories to the PRD
ralph add -t "Add login form" -p 1 -a "Email field validates" -a "Password field works"
ralph add -t "Add authentication" -p 2 -a "JWT tokens generated" -a "Session persists"

# Check status
ralph status

# Run the loop
ralph run
```

## How It Works

Ralph implements a simple but powerful pattern:

1. Read the PRD and find the highest priority pending story
2. Execute your AI agent with a prompt containing the PRD and progress log
3. Agent implements the story, runs tests, commits, and marks it done
4. Check if all stories are complete
5. Repeat until done or max iterations reached

Memory persists between iterations through:
- **Git commits** - Each story = one commit
- **progress.txt** - Learnings and patterns discovered
- **prd.json** - Task status tracking

## Commands

| Command | Description |
|---------|-------------|
| `ralph init` | Initialize Ralph in your project |
| `ralph status` | Show PRD status with progress bar |
| `ralph add` | Add a user story (interactive or via flags) |
| `ralph edit <id>` | Edit an existing story |
| `ralph done <id>` | Mark a story as complete |
| `ralph reset <id>` | Reset a story to pending |
| `ralph delete <id>` | Delete a story |
| `ralph prompt` | View/edit/render the prompt template |
| `ralph log` | View/edit the progress log |
| `ralph run` | Start the Ralph loop |
| `ralph version` | Print version information |

## Configuration

Ralph supports configuration via file, environment variables, and CLI flags (in order of precedence: flags > env > file).

### Config File (ralph.yaml)

```yaml
agent:
  type: claude-code  # claude-code, amp, opencode, codex, custom
  command: ""        # custom command (only if type: custom)
  flags: []          # additional flags
  timeout: 30m       # max time per iteration

loop:
  maxIterations: 25
  sleepBetween: 2s
  stopOnFirstFailure: false

paths:
  prd: .ralph/prd.json
  progress: .ralph/progress.txt
  prompt: .ralph/prompt.md

hooks:
  enabled: true
  onStart: []
  onIteration: []
  onComplete: []
  onFailure: []
```

### Environment Variables

All config options can be set via environment variables with the `RALPH_` prefix:

```bash
export RALPH_AGENT_TYPE=claude-code
export RALPH_LOOP_MAXITERATIONS=50
```

### CLI Flags

```bash
ralph run --max-iterations 10 --agent claude-code
```

## Supported Agents

| Agent | Command |
|-------|---------|
| claude-code | `claude --dangerously-skip-permissions` |
| amp | `amp --dangerously-allow-all` |
| opencode | `opencode` |
| codex | `codex` |
| custom | User-defined command |

## PRD Format

The PRD (Product Requirements Document) is a JSON file containing user stories:

```json
{
  "branchName": "ralph/feature",
  "userStories": [
    {
      "id": "US-001",
      "title": "Add login form",
      "description": "Create a login form with email and password fields",
      "acceptanceCriteria": [
        "Email field with validation",
        "Password field with show/hide toggle",
        "Submit button disabled until valid",
        "typecheck passes",
        "tests pass"
      ],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

## Human-in-the-Loop Mode

For more control, run Ralph one iteration at a time:

```bash
ralph run --once
```

This lets you review each change before continuing.

## Dry Run

See what Ralph would do without executing:

```bash
ralph run --dry-run
```

## Lifecycle Hooks

Run custom commands at different stages:

```yaml
hooks:
  enabled: true
  onStart:
    - "git checkout -b ralph/feature"
  onIteration:
    - "npm run lint"
  onComplete:
    - "./notify.sh 'Ralph finished!'"
  onFailure:
    - "./notify.sh 'Ralph failed'"
```

Hook environment variables:
- `RALPH_ITERATION` - Current iteration number
- `RALPH_STORY_ID` - Current story ID
- `RALPH_HOOK` - Hook type being executed

## Best Practices

### Small Stories

Stories should fit in a single context window. Break large features into smaller pieces:

```
Bad:  "Build entire auth system"
Good: "Add login form"
      "Add email validation"
      "Add auth server action"
      "Add session management"
```

### Explicit Acceptance Criteria

Be specific and testable:

```
Bad:  "Users can log in"
Good: "Email field validates format"
      "Password requires 8+ characters"
      "Error message shows on invalid credentials"
      "typecheck passes"
      "tests pass"
```

### Feedback Loops

Always include verification criteria:
- `typecheck passes`
- `tests pass`
- `npm run lint passes`

### Learnings Compound

Ralph appends learnings to progress.txt. By story 10, it knows patterns from stories 1-9.

## File Structure

After `ralph init`:

```
your-project/
├── ralph.yaml           # Configuration
└── .ralph/
    ├── prd.json         # User stories
    ├── progress.txt     # Progress log
    └── prompt.md        # Agent prompt template
```

## Stop Condition

Ralph stops when:
- All stories have `passes: true`
- Max iterations reached
- Agent outputs `<promise>COMPLETE</promise>`

## License

MIT
