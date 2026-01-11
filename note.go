package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NoteTime is a custom time type that handles the "2006-01-02 15:04" format
type NoteTime struct {
	time.Time
}

const noteTimeFormat = "2006-01-02 15:04"

func (t *NoteTime) UnmarshalYAML(node *yaml.Node) error {
	value := node.Value
	if value == "" {
		t.Time = time.Time{}
		return nil
	}

	// Try custom format first
	parsed, err := time.Parse(noteTimeFormat, value)
	if err == nil {
		t.Time = parsed
		return nil
	}

	// Try RFC3339
	parsed, err = time.Parse(time.RFC3339, value)
	if err == nil {
		t.Time = parsed
		return nil
	}

	// Try other common formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		parsed, err = time.Parse(f, value)
		if err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("cannot parse time: %s", value)
}

func (t NoteTime) MarshalYAML() (interface{}, error) {
	return t.Format(noteTimeFormat), nil
}

// Frontmatter represents the YAML frontmatter of a note
type Frontmatter struct {
	Created NoteTime `yaml:"created"`
	Tags    []string `yaml:"tags"`
	Summary string   `yaml:"summary"`
	Related []string `yaml:"related"`
}

// Note represents a complete note with frontmatter and content
type Note struct {
	Filename    string
	Frontmatter Frontmatter
	Content     string // Body content without frontmatter
}

// ParseNote reads a note file and parses its frontmatter and content
func ParseNote(filepath string) (*Note, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return ParseNoteContent(filepath, data)
}

// ParseNoteContent parses note content from bytes
func ParseNoteContent(filename string, data []byte) (*Note, error) {
	content := string(data)

	// Check for frontmatter
	if !strings.HasPrefix(content, "---\n") {
		// No frontmatter, treat entire content as body
		return &Note{
			Filename: filename,
			Content:  content,
		}, nil
	}

	// Find the closing ---
	rest := content[4:] // Skip opening ---\n
	idx := strings.Index(rest, "\n---\n")
	if idx == -1 {
		// Check for --- at end of file
		if strings.HasSuffix(rest, "\n---") {
			idx = len(rest) - 4
		} else {
			// No closing ---, treat as no frontmatter
			return &Note{
				Filename: filename,
				Content:  content,
			}, nil
		}
	}

	fmContent := rest[:idx]
	body := ""
	if idx+5 < len(rest) {
		body = rest[idx+5:] // Skip \n---\n
	} else if strings.HasSuffix(rest, "\n---") {
		body = ""
	}

	// Parse YAML frontmatter
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(fmContent), &fm); err != nil {
		return nil, fmt.Errorf("invalid frontmatter: %w", err)
	}

	return &Note{
		Filename:    filename,
		Frontmatter: fm,
		Content:     body,
	}, nil
}

// ContentHash computes SHA256 hash of the note content (excluding frontmatter)
// Returns first 12 hex characters
func (n *Note) ContentHash() string {
	hash := sha256.Sum256([]byte(n.Content))
	return hex.EncodeToString(hash[:])[:12]
}

// ToMarkdown renders the note as markdown with frontmatter
func (n *Note) ToMarkdown() string {
	var buf bytes.Buffer

	buf.WriteString("---\n")

	// Format created time
	created := n.Frontmatter.Created.Format(noteTimeFormat)

	// Build YAML manually to control formatting
	buf.WriteString(fmt.Sprintf("created: %s\n", created))

	// Tags
	if len(n.Frontmatter.Tags) == 0 {
		buf.WriteString("tags: []\n")
	} else {
		buf.WriteString("tags: [")
		for i, tag := range n.Frontmatter.Tags {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(tag)
		}
		buf.WriteString("]\n")
	}

	// Summary
	if n.Frontmatter.Summary == "" {
		buf.WriteString("summary: \"\"\n")
	} else {
		buf.WriteString(fmt.Sprintf("summary: %q\n", n.Frontmatter.Summary))
	}

	// Related
	if len(n.Frontmatter.Related) == 0 {
		buf.WriteString("related: []\n")
	} else {
		buf.WriteString("related: [")
		for i, rel := range n.Frontmatter.Related {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(rel)
		}
		buf.WriteString("]\n")
	}

	buf.WriteString("---\n")
	buf.WriteString(n.Content)

	return buf.String()
}

// SaveNote writes a note to the specified path
func (n *Note) Save(filepath string) error {
	return os.WriteFile(filepath, []byte(n.ToMarkdown()), 0644)
}

// UpdateFrontmatter updates the frontmatter of a note file in place
func UpdateFrontmatter(filepath string, tags []string, summary string, related []string) error {
	note, err := ParseNote(filepath)
	if err != nil {
		return err
	}

	if tags != nil {
		note.Frontmatter.Tags = tags
	}
	if summary != "" {
		note.Frontmatter.Summary = summary
	}
	if related != nil {
		note.Frontmatter.Related = related
	}

	return note.Save(filepath)
}

// GetSummaryOrFirstLine returns the summary if available, or the first line truncated
func (n *Note) GetSummaryOrFirstLine() string {
	if n.Frontmatter.Summary != "" {
		return n.Frontmatter.Summary
	}

	// Get first non-empty line
	scanner := bufio.NewScanner(strings.NewReader(n.Content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			if len(line) > 60 {
				return line[:57] + "..."
			}
			return line
		}
	}

	return "(empty)"
}
