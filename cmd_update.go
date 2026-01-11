package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CmdUpdate implements the 'notes update <filename>' command
// Updates note metadata in both frontmatter and .meta.json
func CmdUpdate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: notes update <filename> --tags \"a,b,c\" --summary \"...\" --related \"file1.md,file2.md\"")
	}

	// Extract filename first (may be first arg or after flags)
	var filename string
	var flagArgs []string

	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && filename == "" {
			filename = arg
			// Remaining args are flags
			flagArgs = append(args[:i], args[i+1:]...)
			break
		}
	}

	if filename == "" {
		return fmt.Errorf("usage: notes update <filename> --tags \"a,b,c\" --summary \"...\" --related \"file1.md,file2.md\"")
	}

	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	tagsFlag := fs.String("tags", "", "tags (comma-separated)")
	summaryFlag := fs.String("summary", "", "summary")
	relatedFlag := fs.String("related", "", "related files (comma-separated)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	filename = NormalizeFilename(filename)
	notePath := filepath.Join(notesDir, filename)

	// Check if file exists
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("note not found: %s", filename)
	}

	// Load current note
	note, err := ParseNote(notePath)
	if err != nil {
		return fmt.Errorf("failed to parse note: %w", err)
	}

	// Load meta file
	meta, err := LoadMetaFile(notesDir)
	if err != nil {
		return fmt.Errorf("failed to load meta file: %w", err)
	}

	// Get previous related for bidirectional update
	var prevRelated []string
	if fileMeta := meta.GetFileMeta(filename); fileMeta != nil {
		prevRelated = fileMeta.Related
	} else {
		prevRelated = note.Frontmatter.Related
	}

	// Update tags if provided
	if *tagsFlag != "" {
		tags := parseCSV(*tagsFlag)
		note.Frontmatter.Tags = tags
	}

	// Update summary if provided
	if *summaryFlag != "" {
		note.Frontmatter.Summary = *summaryFlag
	}

	// Update related if provided
	var newRelated []string
	if *relatedFlag != "" {
		newRelated = parseCSV(*relatedFlag)
		// Normalize filenames
		for i := range newRelated {
			newRelated[i] = NormalizeFilename(newRelated[i])
		}
		note.Frontmatter.Related = newRelated
	}

	// Save note with updated frontmatter
	if err := note.Save(notePath); err != nil {
		return fmt.Errorf("failed to save note: %w", err)
	}

	// Update meta file
	fileMeta := meta.GetFileMeta(filename)
	if fileMeta == nil {
		fileMeta = &FileMeta{}
		meta.SetFileMeta(filename, fileMeta)
	}

	fileMeta.ContentHash = note.ContentHash()
	fileMeta.EnrichedAt = time.Now()
	fileMeta.Tags = note.Frontmatter.Tags
	fileMeta.Summary = note.Frontmatter.Summary
	fileMeta.Related = note.Frontmatter.Related

	// Handle bidirectional relations
	if *relatedFlag != "" {
		// Remove old relations that are no longer present
		for _, oldRel := range prevRelated {
			if !contains(newRelated, oldRel) {
				// Remove reverse relation
				if relMeta := meta.GetFileMeta(oldRel); relMeta != nil {
					relMeta.Related = removeString(relMeta.Related, filename)
					// Also update the file's frontmatter
					updateRelatedInFile(notesDir, oldRel, relMeta.Related)
				}
			}
		}

		// Add new relations
		for _, newRel := range newRelated {
			if !contains(prevRelated, newRel) {
				// Add reverse relation
				if relMeta := meta.GetFileMeta(newRel); relMeta != nil {
					if !contains(relMeta.Related, filename) {
						relMeta.Related = append(relMeta.Related, filename)
						// Also update the file's frontmatter
						updateRelatedInFile(notesDir, newRel, relMeta.Related)
					}
				} else {
					// Related file not in meta yet, try to update its frontmatter directly
					relPath := filepath.Join(notesDir, newRel)
					if _, err := os.Stat(relPath); err == nil {
						if relNote, err := ParseNote(relPath); err == nil {
							if !contains(relNote.Frontmatter.Related, filename) {
								relNote.Frontmatter.Related = append(relNote.Frontmatter.Related, filename)
								relNote.Save(relPath)
							}
						}
					}
				}
			}
		}
	}

	// Save meta file
	if err := meta.Save(notesDir); err != nil {
		return fmt.Errorf("failed to save meta file: %w", err)
	}

	fmt.Printf("Updated %s\n", filename)
	return nil
}

func parseCSV(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func updateRelatedInFile(notesDir, filename string, related []string) error {
	notePath := filepath.Join(notesDir, filename)
	note, err := ParseNote(notePath)
	if err != nil {
		return err
	}
	note.Frontmatter.Related = related
	return note.Save(notePath)
}
