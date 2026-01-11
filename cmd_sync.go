package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CmdSync implements the 'notes sync' command
// Rebuilds .meta.json from frontmatter in all note files
func CmdSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	dryRunFlag := fs.Bool("dry-run", false, "show what would change without writing")
	forceFlag := fs.Bool("force", false, "rebuild entire .meta.json from scratch")

	if err := fs.Parse(args); err != nil {
		return err
	}

	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	// Load existing meta or create new one
	var meta *MetaFile
	if *forceFlag {
		meta = &MetaFile{Files: make(map[string]*FileMeta)}
	} else {
		meta, err = LoadMetaFile(notesDir)
		if err != nil {
			return fmt.Errorf("failed to load meta file: %w", err)
		}
	}

	// Find all .md files
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return fmt.Errorf("failed to read notes directory: %w", err)
	}

	var totalCount, updatedCount int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		totalCount++
		filename := entry.Name()
		notePath := filepath.Join(notesDir, filename)

		note, err := ParseNote(notePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", filename, err)
			continue
		}

		existingMeta := meta.GetFileMeta(filename)
		newHash := note.ContentHash()

		// Check what changed
		changes := detectChanges(existingMeta, note, newHash)

		if len(changes) > 0 {
			updatedCount++
			if *dryRunFlag {
				fmt.Printf("Would update: %s (%s)\n", filename, strings.Join(changes, ", "))
			} else {
				fmt.Printf("Updated: %s (%s)\n", filename, strings.Join(changes, ", "))
			}
		}

		if !*dryRunFlag {
			// Update meta
			if existingMeta == nil {
				existingMeta = &FileMeta{}
				meta.SetFileMeta(filename, existingMeta)
			}

			existingMeta.ContentHash = newHash
			existingMeta.Tags = note.Frontmatter.Tags
			existingMeta.Summary = note.Frontmatter.Summary
			existingMeta.Related = note.Frontmatter.Related
			// Preserve enriched_at timestamp
		}
	}

	// Remove entries for files that no longer exist
	for filename := range meta.Files {
		notePath := filepath.Join(notesDir, filename)
		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			if *dryRunFlag {
				fmt.Printf("Would remove: %s (file deleted)\n", filename)
			} else {
				fmt.Printf("Removed: %s (file deleted)\n", filename)
				delete(meta.Files, filename)
			}
		}
	}

	if !*dryRunFlag {
		if err := meta.Save(notesDir); err != nil {
			return fmt.Errorf("failed to save meta file: %w", err)
		}
	}

	unchangedCount := totalCount - updatedCount
	if *dryRunFlag {
		fmt.Printf("\nDry run: would sync %d notes (%d to update, %d unchanged)\n", totalCount, updatedCount, unchangedCount)
	} else {
		fmt.Printf("\nSynced %d notes (%d updated, %d unchanged)\n", totalCount, updatedCount, unchangedCount)
	}

	return nil
}

func detectChanges(existing *FileMeta, note *Note, newHash string) []string {
	var changes []string

	if existing == nil {
		return []string{"new"}
	}

	if existing.ContentHash != newHash {
		changes = append(changes, "content changed")
	}

	if !stringSliceEqual(existing.Tags, note.Frontmatter.Tags) {
		changes = append(changes, "tags changed")
	}

	if existing.Summary != note.Frontmatter.Summary {
		changes = append(changes, "summary changed")
	}

	if !stringSliceEqual(existing.Related, note.Frontmatter.Related) {
		changes = append(changes, "related changed")
	}

	return changes
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
