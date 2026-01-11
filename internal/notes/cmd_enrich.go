package notes

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

	notesList, err := GetNotesNeedingEnrichment(notesDir)
	if err != nil {
		return fmt.Errorf("failed to get notes needing enrichment: %w", err)
	}

	if len(notesList) == 0 {
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
	fmt.Println("## Available CLI Commands")
	fmt.Println()
	fmt.Println("Use these commands to explore notes and find relationships:")
	fmt.Println()
	fmt.Println("- `notes list` - List all notes (newest first) to see what's available")
	fmt.Println("- `notes show <filename>` - Read the full content of any note")
	fmt.Println("- `notes meta <filename>` - View a note's metadata (tags, summary, related) as JSON")
	fmt.Println("- `notes tags` - List all tags with counts to find thematic connections")
	fmt.Println("- `notes graph [filename]` - Show relationship graph (all notes or specific note)")
	fmt.Println("- `notes update <filename>` - Update a note's metadata (see below)")
	fmt.Println()
	fmt.Println("## Finding Related Notes")
	fmt.Println()
	fmt.Println("To identify meaningful relationships between notes:")
	fmt.Println()
	fmt.Println("1. **Browse by tags**: Run `notes tags` to see common themes, then explore notes sharing tags")
	fmt.Println("2. **Read full content**: Use `notes show <filename>` to read notes that might be related")
	fmt.Println("3. **Check existing relationships**: Use `notes graph` to see how notes are already connected")
	fmt.Println("4. **Look for**: shared concepts, references to the same topics, sequential ideas, or complementary information")
	fmt.Println()
	fmt.Println("## Instructions")
	fmt.Println()
	fmt.Println("For each note below:")
	fmt.Println("1. **Tags**: Add 2-5 relevant tags (lowercase, single words or hyphenated)")
	fmt.Println("2. **Summary**: Write a concise one-sentence summary (under 80 chars)")
	fmt.Println("3. **Related**: Identify related notes by exploring the existing notes")
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
	fmt.Println("Use `notes show <filename>` to read each note's content:")
	fmt.Println()
	for _, note := range notesList {
		filename := filepath.Base(note.Filename)
		fmt.Printf("- %s (created: %s)\n", filename, note.Frontmatter.Created.Format("2006-01-02 15:04"))
	}

	return nil
}
