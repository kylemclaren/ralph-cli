package prompt

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/kylemclaren/ralph/internal/prd"
	"github.com/kylemclaren/ralph/internal/progress"
)

// TemplateData holds data passed to the prompt template
type TemplateData struct {
	PRD            string // Full PRD JSON
	Progress       string // Contents of progress.txt
	BranchName     string
	PendingCount   int
	CompletedCount int
	TotalCount     int
	NextStory      *prd.UserStory
}

// Load reads a prompt template from file
func Load(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}
	return string(data), nil
}

// Save writes a prompt template to file
func Save(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write prompt file: %w", err)
	}
	return nil
}

// Render renders the prompt template with the given data
func Render(templateContent string, data TemplateData) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}

	return buf.String(), nil
}

// BuildTemplateData builds template data from PRD and progress files
func BuildTemplateData(p *prd.PRD, prog *progress.Progress) (TemplateData, error) {
	prdJSON, err := p.ToJSON()
	if err != nil {
		return TemplateData{}, err
	}

	total, completed, pending := p.Stats()

	return TemplateData{
		PRD:            prdJSON,
		Progress:       prog.Content,
		BranchName:     p.BranchName,
		PendingCount:   pending,
		CompletedCount: completed,
		TotalCount:     total,
		NextStory:      p.NextStory(),
	}, nil
}

// DefaultPrompt returns the default Ralph prompt template
func DefaultPrompt() string {
	return `# Ralph Agent Instructions

You are Ralph, an autonomous coding agent working through a PRD (Product Requirements Document).

## Your Task

1. Read the PRD below and identify the highest priority story where ` + "`passes: false`" + `
2. Read the progress log for context and patterns from previous work
3. Check you're on the correct branch: ` + "`{{.BranchName}}`" + `
4. Implement that ONE story completely
5. Run typecheck and tests to verify your work
6. Update any AGENTS.md files with learnings if you discovered reusable patterns
7. Commit your changes: ` + "`feat: [ID] - [Title]`" + `
8. Update prd.json: set ` + "`passes: true`" + ` for the completed story
9. Append your learnings to progress.txt

## Current Status

- **Total Stories:** {{.TotalCount}}
- **Completed:** {{.CompletedCount}}
- **Pending:** {{.PendingCount}}
- **Branch:** {{.BranchName}}

## PRD (prd.json)

` + "```json" + `
{{.PRD}}
` + "```" + `

## Progress Log (progress.txt)

` + "```" + `
{{.Progress}}
` + "```" + `

## Progress Format

When appending to progress.txt, use this format:

` + "```" + `
## [Date] - [Story ID]
**[Title]**

Files changed:
- file1.go
- file2.go

**Learnings:**
- Pattern discovered
- Gotcha encountered
---
` + "```" + `

## Codebase Patterns

Add reusable patterns to the TOP of progress.txt under "## Codebase Patterns":
- Migrations: Use IF NOT EXISTS
- React: useRef<Timeout | null>(null)
- Tests: Run with -v flag

## Stop Condition

If ALL stories have ` + "`passes: true`" + `, output exactly:

<promise>COMPLETE</promise>

Otherwise, end your response normally after completing one story.

## Important Rules

1. **One story at a time** - Don't try to do multiple stories
2. **Verify your work** - Run typecheck and tests before committing
3. **Commit after each story** - Each story = one commit
4. **Update the PRD** - Mark the story as passing when done
5. **Log your learnings** - Help future iterations learn from your work
`
}

// Exists checks if the prompt file exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Create creates a new prompt file with default content
func Create(path string) error {
	return Save(path, DefaultPrompt())
}
