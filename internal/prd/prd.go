package prd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// PRD represents the Product Requirements Document
type PRD struct {
	BranchName  string      `json:"branchName"`
	UserStories []UserStory `json:"userStories"`
}

// UserStory represents a single user story/task
type UserStory struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description,omitempty"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Priority           int      `json:"priority"`
	Passes             bool     `json:"passes"`
	Notes              string   `json:"notes,omitempty"`
}

// Load reads a PRD from a JSON file
func Load(path string) (*PRD, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read PRD file: %w", err)
	}

	var prd PRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return nil, fmt.Errorf("failed to parse PRD JSON: %w", err)
	}

	return &prd, nil
}

// Save writes the PRD to a JSON file
func (p *PRD) Save(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PRD: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write PRD file: %w", err)
	}

	return nil
}

// NewPRD creates a new empty PRD
func NewPRD(branchName string) *PRD {
	return &PRD{
		BranchName:  branchName,
		UserStories: []UserStory{},
	}
}

// AddStory adds a new user story to the PRD
func (p *PRD) AddStory(story UserStory) {
	// Generate ID if not provided
	if story.ID == "" {
		story.ID = p.generateID()
	}
	p.UserStories = append(p.UserStories, story)
}

// GetStory returns a story by ID
func (p *PRD) GetStory(id string) *UserStory {
	id = strings.ToUpper(id)
	for i := range p.UserStories {
		if strings.ToUpper(p.UserStories[i].ID) == id {
			return &p.UserStories[i]
		}
	}
	return nil
}

// UpdateStory updates an existing story
func (p *PRD) UpdateStory(story UserStory) error {
	for i := range p.UserStories {
		if strings.EqualFold(p.UserStories[i].ID, story.ID) {
			p.UserStories[i] = story
			return nil
		}
	}
	return fmt.Errorf("story %s not found", story.ID)
}

// MarkDone marks a story as passing
func (p *PRD) MarkDone(id string) error {
	story := p.GetStory(id)
	if story == nil {
		return fmt.Errorf("story %s not found", id)
	}
	story.Passes = true
	return nil
}

// MarkPending marks a story as not passing
func (p *PRD) MarkPending(id string) error {
	story := p.GetStory(id)
	if story == nil {
		return fmt.Errorf("story %s not found", id)
	}
	story.Passes = false
	return nil
}

// DeleteStory removes a story by ID
func (p *PRD) DeleteStory(id string) error {
	id = strings.ToUpper(id)
	for i := range p.UserStories {
		if strings.ToUpper(p.UserStories[i].ID) == id {
			p.UserStories = append(p.UserStories[:i], p.UserStories[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("story %s not found", id)
}

// PendingStories returns all stories that haven't passed yet
func (p *PRD) PendingStories() []UserStory {
	var pending []UserStory
	for _, s := range p.UserStories {
		if !s.Passes {
			pending = append(pending, s)
		}
	}
	return pending
}

// CompletedStories returns all stories that have passed
func (p *PRD) CompletedStories() []UserStory {
	var completed []UserStory
	for _, s := range p.UserStories {
		if s.Passes {
			completed = append(completed, s)
		}
	}
	return completed
}

// NextStory returns the highest priority pending story
func (p *PRD) NextStory() *UserStory {
	pending := p.PendingStories()
	if len(pending) == 0 {
		return nil
	}

	// Sort by priority (lower number = higher priority)
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Priority < pending[j].Priority
	})

	return &pending[0]
}

// IsComplete returns true if all stories pass
func (p *PRD) IsComplete() bool {
	for _, s := range p.UserStories {
		if !s.Passes {
			return false
		}
	}
	return len(p.UserStories) > 0
}

// Stats returns statistics about the PRD
func (p *PRD) Stats() (total, completed, pending int) {
	total = len(p.UserStories)
	for _, s := range p.UserStories {
		if s.Passes {
			completed++
		} else {
			pending++
		}
	}
	return
}

// generateID generates a new story ID
func (p *PRD) generateID() string {
	maxNum := 0
	for _, s := range p.UserStories {
		var num int
		if _, err := fmt.Sscanf(s.ID, "US-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}
	return fmt.Sprintf("US-%03d", maxNum+1)
}

// ToJSON returns the PRD as a formatted JSON string
func (p *PRD) ToJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DefaultPRD returns a PRD with example content
func DefaultPRD() *PRD {
	return &PRD{
		BranchName: "ralph/feature",
		UserStories: []UserStory{
			{
				ID:          "US-001",
				Title:       "Example user story",
				Description: "Describe what this story accomplishes",
				AcceptanceCriteria: []string{
					"First acceptance criterion",
					"Second acceptance criterion",
					"typecheck passes",
					"tests pass",
				},
				Priority: 1,
				Passes:   false,
				Notes:    "",
			},
		},
	}
}

// FormatStoryForDisplay formats a story for terminal display
func (s *UserStory) FormatForDisplay() string {
	status := "[ ]"
	if s.Passes {
		status = "[x]"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s: %s (P%d)\n", status, s.ID, s.Title, s.Priority))

	if s.Description != "" {
		sb.WriteString(fmt.Sprintf("    %s\n", s.Description))
	}

	if len(s.AcceptanceCriteria) > 0 {
		sb.WriteString("    Acceptance Criteria:\n")
		for _, ac := range s.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("      - %s\n", ac))
		}
	}

	if s.Notes != "" {
		sb.WriteString(fmt.Sprintf("    Notes: %s\n", s.Notes))
	}

	return sb.String()
}

// CreatedAt returns a formatted timestamp for logs
func CreatedAt() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
