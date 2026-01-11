package notes

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMetaFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "notes-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test loading non-existent file returns empty meta
	meta, err := LoadMetaFile(tmpDir)
	if err != nil {
		t.Fatalf("LoadMetaFile() error = %v", err)
	}
	if meta.Files == nil {
		t.Error("Files map should be initialized")
	}
	if len(meta.Files) != 0 {
		t.Error("Files map should be empty")
	}
}

func TestMetaFileSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "notes-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and save meta
	meta := &MetaFile{
		Files: map[string]*FileMeta{
			"test.md": {
				ContentHash: "abc123def456",
				EnrichedAt:  time.Now(),
				Tags:        []string{"tag1", "tag2"},
				Summary:     "Test summary",
				Related:     []string{"other.md"},
			},
		},
	}

	err = meta.Save(tmpDir)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check file exists
	metaPath := filepath.Join(tmpDir, ".meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("Meta file should exist")
	}

	// Load and verify
	loaded, err := LoadMetaFile(tmpDir)
	if err != nil {
		t.Fatalf("LoadMetaFile() error = %v", err)
	}

	fileMeta := loaded.GetFileMeta("test.md")
	if fileMeta == nil {
		t.Fatal("Should find test.md in loaded meta")
	}
	if fileMeta.ContentHash != "abc123def456" {
		t.Errorf("ContentHash = %q, want %q", fileMeta.ContentHash, "abc123def456")
	}
	if fileMeta.Summary != "Test summary" {
		t.Errorf("Summary = %q, want %q", fileMeta.Summary, "Test summary")
	}
	if len(fileMeta.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(fileMeta.Tags))
	}
}

func TestNeedsEnrichment(t *testing.T) {
	meta := &MetaFile{
		Files: map[string]*FileMeta{
			"enriched.md": {
				ContentHash: "abc123",
			},
		},
	}

	// File not in meta needs enrichment
	if !meta.NeedsEnrichment("new.md", "xyz789") {
		t.Error("New file should need enrichment")
	}

	// File with same hash doesn't need enrichment
	if meta.NeedsEnrichment("enriched.md", "abc123") {
		t.Error("File with same hash should not need enrichment")
	}

	// File with different hash needs enrichment
	if !meta.NeedsEnrichment("enriched.md", "different") {
		t.Error("File with different hash should need enrichment")
	}
}

func TestBidirectionalRelations(t *testing.T) {
	meta := &MetaFile{
		Files: map[string]*FileMeta{
			"a.md": {Related: []string{}},
			"b.md": {Related: []string{}},
		},
	}

	// Add relation
	meta.AddRelation("a.md", "b.md")

	aMeta := meta.GetFileMeta("a.md")
	bMeta := meta.GetFileMeta("b.md")

	if !Contains(aMeta.Related, "b.md") {
		t.Error("a.md should be related to b.md")
	}
	if !Contains(bMeta.Related, "a.md") {
		t.Error("b.md should be related to a.md")
	}

	// Remove relation
	meta.RemoveRelation("a.md", "b.md")

	if Contains(aMeta.Related, "b.md") {
		t.Error("a.md should not be related to b.md after removal")
	}
	if Contains(bMeta.Related, "a.md") {
		t.Error("b.md should not be related to a.md after removal")
	}
}

func TestUpdateFromNote(t *testing.T) {
	meta := &MetaFile{
		Files: make(map[string]*FileMeta),
	}

	note := &Note{
		Filename: "test.md",
		Frontmatter: Frontmatter{
			Tags:    []string{"tag1", "tag2"},
			Summary: "Test summary",
			Related: []string{"other.md"},
		},
		Content: "Body content",
	}

	meta.UpdateFromNote(note)

	fileMeta := meta.GetFileMeta("test.md")
	if fileMeta == nil {
		t.Fatal("Should have created meta entry")
	}
	if fileMeta.Summary != "Test summary" {
		t.Errorf("Summary = %q, want %q", fileMeta.Summary, "Test summary")
	}
	if len(fileMeta.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(fileMeta.Tags))
	}
	if fileMeta.ContentHash == "" {
		t.Error("ContentHash should be set")
	}
}
