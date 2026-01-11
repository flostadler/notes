package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CmdEnrich implements the 'notes enrich' command
// Outputs structured prompt for AI enrichment
func CmdEnrich(args []string) error {
	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	notes, err := GetNotesNeedingEnrichment(notesDir)
	if err != nil {
		return fmt.Errorf("failed to get notes needing enrichment: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("All notes up to date")
		return nil
	}

	// Load meta to get existing notes for relation suggestions
	meta, err := LoadMetaFile(notesDir)
	if err != nil {
		return fmt.Errorf("failed to load meta file: %w", err)
	}

	// Build context of existing enriched notes
	var existingNotes []string
	for filename, fileMeta := range meta.Files {
		if fileMeta.Summary != "" {
			existingNotes = append(existingNotes, fmt.Sprintf("- %s: %s (tags: %s)",
				filename, fileMeta.Summary, strings.Join(fileMeta.Tags, ", ")))
		}
	}

	// Output the prompt
	fmt.Println("# Notes Enrichment Request")
	fmt.Println()
	fmt.Println("Please enrich the following notes by adding tags, a summary, and identifying related notes.")
	fmt.Println()
	fmt.Println("## Instructions")
	fmt.Println()
	fmt.Println("For each note below:")
	fmt.Println("1. **Tags**: Add 2-5 relevant tags (lowercase, single words or hyphenated)")
	fmt.Println("2. **Summary**: Write a concise one-sentence summary (under 80 chars)")
	fmt.Println("3. **Related**: Identify related notes from the existing notes list")
	fmt.Println()
	fmt.Println("After analyzing, use the `notes update` command for each note:")
	fmt.Println("```")
	fmt.Println("notes update <filename> --tags \"tag1,tag2,tag3\" --summary \"Your summary here\" --related \"file1.md,file2.md\"")
	fmt.Println("```")
	fmt.Println()

	if len(existingNotes) > 0 {
		fmt.Println("## Existing Notes (for finding relations)")
		fmt.Println()
		for _, note := range existingNotes {
			fmt.Println(note)
		}
		fmt.Println()
	}

	fmt.Println("## Notes to Enrich")
	fmt.Println()

	for _, note := range notes {
		filename := filepath.Base(note.Filename)
		fmt.Printf("### %s\n", filename)
		fmt.Printf("Created: %s\n", note.Frontmatter.Created.Format("2006-01-02 15:04"))
		fmt.Println()
		fmt.Println("```")
		fmt.Print(strings.TrimSpace(note.Content))
		fmt.Println()
		fmt.Println("```")
		fmt.Println()
	}

	return nil
}
