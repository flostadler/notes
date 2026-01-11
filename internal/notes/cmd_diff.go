package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CmdDiff implements the 'notes diff' command
// Lists notes that need enrichment
func CmdDiff(args []string) error {
	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	meta, err := LoadMetaFile(notesDir)
	if err != nil {
		return fmt.Errorf("failed to load meta file: %w", err)
	}

	// Find all .md files
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return fmt.Errorf("failed to read notes directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		notePath := filepath.Join(notesDir, entry.Name())
		note, err := ParseNote(notePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		currentHash := note.ContentHash()
		if meta.NeedsEnrichment(entry.Name(), currentHash) {
			fmt.Println(entry.Name())
		}
	}

	return nil
}

// GetNotesNeedingEnrichment returns a list of notes that need enrichment
func GetNotesNeedingEnrichment(notesDir string) ([]*Note, error) {
	meta, err := LoadMetaFile(notesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load meta file: %w", err)
	}

	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read notes directory: %w", err)
	}

	var notesList []*Note
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		notePath := filepath.Join(notesDir, entry.Name())
		note, err := ParseNote(notePath)
		if err != nil {
			continue
		}

		currentHash := note.ContentHash()
		if meta.NeedsEnrichment(entry.Name(), currentHash) {
			notesList = append(notesList, note)
		}
	}

	return notesList, nil
}
