package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestDir(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "notes-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Set NOTES_DIR to the temp directory
	oldNotesDir := os.Getenv("NOTES_DIR")
	os.Setenv("NOTES_DIR", tmpDir)

	cleanup := func() {
		os.Setenv("NOTES_DIR", oldNotesDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func createTestNote(t *testing.T, dir, filename, content string) {
	created, _ := time.Parse("2006-01-02 15:04", "2025-01-11 14:23")
	note := &Note{
		Filename: filename,
		Frontmatter: Frontmatter{
			Created: NoteTime{created},
			Tags:    []string{},
			Related: []string{},
		},
		Content: "\n" + content + "\n",
	}
	err := note.Save(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}
}

func createEnrichedTestNote(t *testing.T, dir, filename, content string, tags []string, summary string) {
	created, _ := time.Parse("2006-01-02 15:04", "2025-01-11 14:23")
	note := &Note{
		Filename: filename,
		Frontmatter: Frontmatter{
			Created: NoteTime{created},
			Tags:    tags,
			Summary: summary,
			Related: []string{},
		},
		Content: "\n" + content + "\n",
	}
	err := note.Save(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	// Add to meta
	meta, _ := LoadMetaFile(dir)
	meta.UpdateFromNoteWithEnrichment(note)
	meta.Save(dir)
}

func TestCmdNewWithContent(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	err := CmdNew([]string{"This is my test note content"})
	if err != nil {
		t.Fatalf("CmdNew() error = %v", err)
	}

	// Check file was created
	entries, _ := os.ReadDir(tmpDir)
	if len(entries) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(entries))
	}

	// Check content
	content, _ := os.ReadFile(filepath.Join(tmpDir, entries[0].Name()))
	if !strings.Contains(string(content), "This is my test note content") {
		t.Error("File should contain the note content")
	}
}

func TestCmdDiff(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create an unenriched note
	createTestNote(t, tmpDir, "2025-01-11-1423.md", "Test content")

	// Create an enriched note
	createEnrichedTestNote(t, tmpDir, "2025-01-11-1424.md", "Enriched content", []string{"tag1"}, "Summary")

	// Run diff - should only show unenriched note
	// Capture output would require refactoring, so we just verify no error
	err := CmdDiff([]string{})
	if err != nil {
		t.Fatalf("CmdDiff() error = %v", err)
	}
}

func TestCmdList(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createEnrichedTestNote(t, tmpDir, "2025-01-11-1423.md", "Content 1", []string{"tag1"}, "Summary 1")
	createEnrichedTestNote(t, tmpDir, "2025-01-11-1424.md", "Content 2", []string{"tag2"}, "Summary 2")

	err := CmdList([]string{})
	if err != nil {
		t.Fatalf("CmdList() error = %v", err)
	}
}

func TestCmdListWithFilters(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createEnrichedTestNote(t, tmpDir, "2025-01-11-1423.md", "Content 1", []string{"neo"}, "Summary 1")
	createEnrichedTestNote(t, tmpDir, "2025-01-11-1424.md", "Content 2", []string{"meeting"}, "Summary 2")

	// Filter by tag
	err := CmdList([]string{"--tags", "neo"})
	if err != nil {
		t.Fatalf("CmdList() with tag filter error = %v", err)
	}

	// With limit
	err = CmdList([]string{"--limit", "1"})
	if err != nil {
		t.Fatalf("CmdList() with limit error = %v", err)
	}

	// Raw output
	err = CmdList([]string{"--raw"})
	if err != nil {
		t.Fatalf("CmdList() with raw error = %v", err)
	}
}

func TestCmdShow(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createTestNote(t, tmpDir, "2025-01-11-1423.md", "This is the note content")

	// With .md extension
	err := CmdShow([]string{"2025-01-11-1423.md"})
	if err != nil {
		t.Fatalf("CmdShow() error = %v", err)
	}

	// Without .md extension
	err = CmdShow([]string{"2025-01-11-1423"})
	if err != nil {
		t.Fatalf("CmdShow() without extension error = %v", err)
	}
}

func TestCmdShowNotFound(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	err := CmdShow([]string{"nonexistent.md"})
	if err == nil {
		t.Error("CmdShow() should error for nonexistent file")
	}
}

func TestCmdMeta(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createEnrichedTestNote(t, tmpDir, "2025-01-11-1423.md", "Content", []string{"tag1", "tag2"}, "Test summary")

	err := CmdMeta([]string{"2025-01-11-1423"})
	if err != nil {
		t.Fatalf("CmdMeta() error = %v", err)
	}
}

func TestCmdUpdate(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createTestNote(t, tmpDir, "2025-01-11-1423.md", "Content")

	err := CmdUpdate([]string{
		"2025-01-11-1423.md",
		"--tags", "neo,eval,idea",
		"--summary", "Test summary here",
		"--related", "other.md",
	})
	if err != nil {
		t.Fatalf("CmdUpdate() error = %v", err)
	}

	// Verify note was updated
	note, err := ParseNote(filepath.Join(tmpDir, "2025-01-11-1423.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(note.Frontmatter.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(note.Frontmatter.Tags))
	}
	if note.Frontmatter.Summary != "Test summary here" {
		t.Errorf("Summary = %q, want %q", note.Frontmatter.Summary, "Test summary here")
	}

	// Verify meta was updated
	meta, _ := LoadMetaFile(tmpDir)
	fileMeta := meta.GetFileMeta("2025-01-11-1423.md")
	if fileMeta == nil {
		t.Fatal("Meta should have entry")
	}
	if fileMeta.Summary != "Test summary here" {
		t.Errorf("Meta summary = %q, want %q", fileMeta.Summary, "Test summary here")
	}
}

func TestCmdUpdateBidirectional(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createTestNote(t, tmpDir, "a.md", "Content A")
	createTestNote(t, tmpDir, "b.md", "Content B")

	// Initialize meta for b.md
	meta, _ := LoadMetaFile(tmpDir)
	note, _ := ParseNote(filepath.Join(tmpDir, "b.md"))
	meta.UpdateFromNote(note)
	meta.Save(tmpDir)

	// Update a.md to relate to b.md
	err := CmdUpdate([]string{
		"a.md",
		"--tags", "test",
		"--summary", "Summary A",
		"--related", "b.md",
	})
	if err != nil {
		t.Fatalf("CmdUpdate() error = %v", err)
	}

	// Check that b.md now relates back to a.md
	meta, _ = LoadMetaFile(tmpDir)
	bMeta := meta.GetFileMeta("b.md")
	if bMeta == nil || !contains(bMeta.Related, "a.md") {
		t.Error("b.md should have bidirectional relation to a.md")
	}
}

func TestCmdSync(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create notes with enriched frontmatter (but no meta)
	created, _ := time.Parse("2006-01-02 15:04", "2025-01-11 14:23")
	note := &Note{
		Filename: "test.md",
		Frontmatter: Frontmatter{
			Created: NoteTime{created},
			Tags:    []string{"tag1", "tag2"},
			Summary: "Manual summary",
			Related: []string{},
		},
		Content: "\nContent here\n",
	}
	note.Save(filepath.Join(tmpDir, "test.md"))

	// Run sync
	err := CmdSync([]string{})
	if err != nil {
		t.Fatalf("CmdSync() error = %v", err)
	}

	// Verify meta was created
	meta, _ := LoadMetaFile(tmpDir)
	fileMeta := meta.GetFileMeta("test.md")
	if fileMeta == nil {
		t.Fatal("Meta should have entry after sync")
	}
	if fileMeta.Summary != "Manual summary" {
		t.Errorf("Summary = %q, want %q", fileMeta.Summary, "Manual summary")
	}
	if len(fileMeta.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(fileMeta.Tags))
	}
}

func TestCmdSyncDryRun(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createTestNote(t, tmpDir, "test.md", "Content")

	err := CmdSync([]string{"--dry-run"})
	if err != nil {
		t.Fatalf("CmdSync() dry run error = %v", err)
	}

	// Meta should NOT be created in dry run
	metaPath := filepath.Join(tmpDir, ".meta.json")
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		// File might exist but be empty
		data, _ := os.ReadFile(metaPath)
		if len(data) > 5 { // more than just "{}"
			t.Error("Meta should not be created in dry run")
		}
	}
}

func TestCmdGraph(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create related notes
	createEnrichedTestNote(t, tmpDir, "a.md", "Content A", []string{"neo"}, "Summary A")
	createEnrichedTestNote(t, tmpDir, "b.md", "Content B", []string{"neo"}, "Summary B")

	// Add relation
	meta, _ := LoadMetaFile(tmpDir)
	meta.AddRelation("a.md", "b.md")
	meta.Save(tmpDir)

	// Test all connections view
	err := CmdGraph([]string{})
	if err != nil {
		t.Fatalf("CmdGraph() error = %v", err)
	}

	// Test specific note view
	err = CmdGraph([]string{"a.md"})
	if err != nil {
		t.Fatalf("CmdGraph(a.md) error = %v", err)
	}

	// Test with JSON output
	err = CmdGraph([]string{"--json"})
	if err != nil {
		t.Fatalf("CmdGraph(--json) error = %v", err)
	}
}

func TestCmdTags(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createEnrichedTestNote(t, tmpDir, "a.md", "Content A", []string{"neo", "eval"}, "Summary A")
	createEnrichedTestNote(t, tmpDir, "b.md", "Content B", []string{"neo", "meeting"}, "Summary B")
	createEnrichedTestNote(t, tmpDir, "c.md", "Content C", []string{"idea"}, "Summary C")

	err := CmdTags([]string{})
	if err != nil {
		t.Fatalf("CmdTags() error = %v", err)
	}
}

func TestCmdEnrich(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createTestNote(t, tmpDir, "unenriched.md", "Content needing enrichment")

	err := CmdEnrich([]string{})
	if err != nil {
		t.Fatalf("CmdEnrich() error = %v", err)
	}
}

func TestCmdEnrichAllUpToDate(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	createEnrichedTestNote(t, tmpDir, "enriched.md", "Already enriched", []string{"tag"}, "Summary")

	err := CmdEnrich([]string{})
	if err != nil {
		t.Fatalf("CmdEnrich() error = %v", err)
	}
}
