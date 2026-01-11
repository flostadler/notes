package notes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MetaOutput represents the JSON output for notes meta command
type MetaOutput struct {
	Created     string   `json:"created"`
	Tags        []string `json:"tags"`
	Summary     string   `json:"summary"`
	Related     []string `json:"related"`
	EnrichedAt  string   `json:"enriched_at,omitempty"`
	ContentHash string   `json:"content_hash"`
	Unenriched  bool     `json:"unenriched,omitempty"`
}

// CmdMeta implements the 'notes meta <filename>' command
// Prints note metadata as JSON
func CmdMeta(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: notes meta <filename>")
	}

	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	filename := NormalizeFilename(args[0])
	notePath := filepath.Join(notesDir, filename)

	// Check if file exists
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("note not found: %s", filename)
	}

	// Try to get from meta file first
	meta, err := LoadMetaFile(notesDir)
	if err != nil {
		return fmt.Errorf("failed to load meta file: %w", err)
	}

	fileMeta := meta.GetFileMeta(filename)
	if fileMeta != nil && fileMeta.ContentHash != "" {
		output := MetaOutput{
			Tags:        fileMeta.Tags,
			Summary:     fileMeta.Summary,
			Related:     fileMeta.Related,
			ContentHash: fileMeta.ContentHash,
		}

		if !fileMeta.EnrichedAt.IsZero() {
			output.EnrichedAt = fileMeta.EnrichedAt.Format("2006-01-02T15:04:05Z")
		}

		// Get created from frontmatter
		note, err := ParseNote(notePath)
		if err == nil {
			output.Created = note.Frontmatter.Created.Format("2006-01-02T15:04:05Z")
		}

		if output.Tags == nil {
			output.Tags = []string{}
		}
		if output.Related == nil {
			output.Related = []string{}
		}

		return outputJSON(output)
	}

	// Not in meta file, parse from frontmatter
	note, err := ParseNote(notePath)
	if err != nil {
		return fmt.Errorf("failed to parse note: %w", err)
	}

	output := MetaOutput{
		Created:     note.Frontmatter.Created.Format("2006-01-02T15:04:05Z"),
		Tags:        note.Frontmatter.Tags,
		Summary:     note.Frontmatter.Summary,
		Related:     note.Frontmatter.Related,
		ContentHash: note.ContentHash(),
		Unenriched:  true,
	}

	if output.Tags == nil {
		output.Tags = []string{}
	}
	if output.Related == nil {
		output.Related = []string{}
	}

	return outputJSON(output)
}

func outputJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
