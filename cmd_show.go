package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// CmdShow implements the 'notes show <filename>' command
// Prints note content without frontmatter
func CmdShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: notes show <filename>")
	}

	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	filename := NormalizeFilename(args[0])
	notePath := filepath.Join(notesDir, filename)

	note, err := ParseNote(notePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("note not found: %s", filename)
		}
		return fmt.Errorf("failed to parse note: %w", err)
	}

	// Print content without leading newline if present
	content := note.Content
	if len(content) > 0 && content[0] == '\n' {
		content = content[1:]
	}
	fmt.Print(content)

	return nil
}
