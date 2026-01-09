package progress

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Progress manages the progress.txt file
type Progress struct {
	Path    string
	Content string
}

// Load reads the progress file
func Load(path string) (*Progress, error) {
	p := &Progress{Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			p.Content = ""
			return p, nil
		}
		return nil, fmt.Errorf("failed to read progress file: %w", err)
	}

	p.Content = string(data)
	return p, nil
}

// Save writes the progress file
func (p *Progress) Save() error {
	if err := os.WriteFile(p.Path, []byte(p.Content), 0644); err != nil {
		return fmt.Errorf("failed to write progress file: %w", err)
	}
	return nil
}

// Append adds content to the end of the progress file
func (p *Progress) Append(content string) {
	if p.Content != "" && !strings.HasSuffix(p.Content, "\n") {
		p.Content += "\n"
	}
	p.Content += content
}

// AppendEntry adds a formatted log entry
func (p *Progress) AppendEntry(storyID, title string, filesChanged []string, learnings []string) {
	entry := formatEntry(storyID, title, filesChanged, learnings)
	p.Append(entry)
}

// formatEntry creates a formatted progress entry
func formatEntry(storyID, title string, filesChanged []string, learnings []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n## %s - %s\n", time.Now().Format("2006-01-02 15:04"), storyID))
	sb.WriteString(fmt.Sprintf("**%s**\n\n", title))

	if len(filesChanged) > 0 {
		sb.WriteString("Files changed:\n")
		for _, f := range filesChanged {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(learnings) > 0 {
		sb.WriteString("**Learnings:**\n")
		for _, l := range learnings {
			sb.WriteString(fmt.Sprintf("- %s\n", l))
		}
	}

	sb.WriteString("\n---\n")
	return sb.String()
}

// GetCodebasePatterns extracts the codebase patterns section
func (p *Progress) GetCodebasePatterns() string {
	// Look for ## Codebase Patterns section
	start := strings.Index(p.Content, "## Codebase Patterns")
	if start == -1 {
		return ""
	}

	// Find the end (next ## or end of file)
	rest := p.Content[start+len("## Codebase Patterns"):]
	end := strings.Index(rest, "\n## ")
	if end == -1 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

// DefaultProgress returns initial progress file content
func DefaultProgress() string {
	return fmt.Sprintf(`# Ralph Progress Log
Started: %s

## Codebase Patterns
<!-- Add reusable patterns discovered during implementation -->

## Key Files
<!-- Document important files for context -->

---
`, time.Now().Format("2006-01-02"))
}

// Create creates a new progress file with default content
func Create(path string) (*Progress, error) {
	p := &Progress{
		Path:    path,
		Content: DefaultProgress(),
	}
	if err := p.Save(); err != nil {
		return nil, err
	}
	return p, nil
}

// Exists checks if the progress file exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
