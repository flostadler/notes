package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseNoteContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantSummary string
		wantTags    []string
		wantContent string
		wantErr     bool
	}{
		{
			name: "valid frontmatter",
			content: `---
created: 2025-01-11 14:23
tags: [neo, eval, idea]
summary: "Test summary"
related: []
---

Body content here.
`,
			wantSummary: "Test summary",
			wantTags:    []string{"neo", "eval", "idea"},
			wantContent: "\nBody content here.\n",
		},
		{
			name: "empty frontmatter",
			content: `---
created: 2025-01-11 14:23
tags: []
summary: ""
related: []
---

Body content here.
`,
			wantSummary: "",
			wantTags:    []string{},
			wantContent: "\nBody content here.\n",
		},
		{
			name:        "no frontmatter",
			content:     "Just plain content\nNo frontmatter here.",
			wantContent: "Just plain content\nNo frontmatter here.",
		},
		{
			name: "frontmatter only",
			content: `---
created: 2025-01-11 14:23
tags: []
summary: ""
related: []
---
`,
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note, err := ParseNoteContent("test.md", []byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNoteContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if note.Frontmatter.Summary != tt.wantSummary {
				t.Errorf("Summary = %q, want %q", note.Frontmatter.Summary, tt.wantSummary)
			}

			if len(tt.wantTags) > 0 {
				if len(note.Frontmatter.Tags) != len(tt.wantTags) {
					t.Errorf("Tags = %v, want %v", note.Frontmatter.Tags, tt.wantTags)
				}
			}

			if note.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q", note.Content, tt.wantContent)
			}
		})
	}
}

func TestContentHash(t *testing.T) {
	note1 := &Note{
		Content: "Some content here",
	}
	note2 := &Note{
		Content: "Some content here",
	}
	note3 := &Note{
		Content: "Different content",
	}

	// Same content should produce same hash
	if note1.ContentHash() != note2.ContentHash() {
		t.Errorf("Same content should produce same hash")
	}

	// Different content should produce different hash
	if note1.ContentHash() == note3.ContentHash() {
		t.Errorf("Different content should produce different hash")
	}

	// Hash should be 12 characters
	if len(note1.ContentHash()) != 12 {
		t.Errorf("Hash should be 12 characters, got %d", len(note1.ContentHash()))
	}
}

func TestContentHashIgnoresFrontmatter(t *testing.T) {
	note1 := &Note{
		Frontmatter: Frontmatter{
			Tags:    []string{"tag1"},
			Summary: "Summary 1",
		},
		Content: "Same content",
	}
	note2 := &Note{
		Frontmatter: Frontmatter{
			Tags:    []string{"different", "tags"},
			Summary: "Different summary",
		},
		Content: "Same content",
	}

	// Hash should be the same since content is the same
	if note1.ContentHash() != note2.ContentHash() {
		t.Errorf("Hash should ignore frontmatter, got %s vs %s", note1.ContentHash(), note2.ContentHash())
	}
}

func TestToMarkdown(t *testing.T) {
	created, _ := time.Parse("2006-01-02 15:04", "2025-01-11 14:23")
	note := &Note{
		Frontmatter: Frontmatter{
			Created: NoteTime{created},
			Tags:    []string{"neo", "eval"},
			Summary: "Test summary",
			Related: []string{"other.md"},
		},
		Content: "\nBody content here.\n",
	}

	markdown := note.ToMarkdown()

	// Check key components are present
	if !strings.Contains(markdown, "created: 2025-01-11 14:23") {
		t.Errorf("Markdown should contain created date, got:\n%s", markdown)
	}
	if !strings.Contains(markdown, "tags: [neo, eval]") {
		t.Errorf("Markdown should contain tags, got:\n%s", markdown)
	}
	if !strings.Contains(markdown, `summary: "Test summary"`) {
		t.Errorf("Markdown should contain summary, got:\n%s", markdown)
	}
	if !strings.Contains(markdown, "related: [other.md]") {
		t.Errorf("Markdown should contain related, got:\n%s", markdown)
	}
	if !strings.Contains(markdown, "Body content here.") {
		t.Errorf("Markdown should contain body content, got:\n%s", markdown)
	}
}

func TestGetSummaryOrFirstLine(t *testing.T) {
	tests := []struct {
		name     string
		note     *Note
		expected string
	}{
		{
			name: "has summary",
			note: &Note{
				Frontmatter: Frontmatter{Summary: "This is the summary"},
				Content:     "First line\nSecond line",
			},
			expected: "This is the summary",
		},
		{
			name: "no summary, uses first line",
			note: &Note{
				Content: "First line of content\nSecond line",
			},
			expected: "First line of content",
		},
		{
			name: "long first line gets truncated",
			note: &Note{
				Content: "This is a very long first line that should be truncated because it exceeds sixty characters in length\nSecond line",
			},
			expected: "This is a very long first line that should be truncated b...",
		},
		{
			name: "empty content",
			note: &Note{
				Content: "",
			},
			expected: "(empty)",
		},
		{
			name: "whitespace only content",
			note: &Note{
				Content: "\n\n   \n",
			},
			expected: "(empty)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.note.GetSummaryOrFirstLine()
			if result != tt.expected {
				t.Errorf("GetSummaryOrFirstLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNormalizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2025-01-11-1423", "2025-01-11-1423.md"},
		{"2025-01-11-1423.md", "2025-01-11-1423.md"},
		{"note", "note.md"},
		{"note.md", "note.md"},
	}

	for _, tt := range tests {
		result := NormalizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateFilename(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "notes-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// First filename should not have suffix
	filename1, err := GenerateFilename(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if filename1[len(filename1)-5:] != "-1.md" && filename1[len(filename1)-3:] != ".md" {
		t.Errorf("Unexpected filename format: %s", filename1)
	}

	// Create that file
	os.WriteFile(filepath.Join(tmpDir, filename1), []byte("test"), 0644)

	// Second filename should have -1 suffix
	filename2, err := GenerateFilename(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if filename1 == filename2 {
		t.Errorf("Second filename should be different from first")
	}
}
