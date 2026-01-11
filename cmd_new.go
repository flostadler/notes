package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CmdNew implements the 'notes new [content]' command
func CmdNew(args []string) error {
	notesDir, err := EnsureNotesDir()
	if err != nil {
		return fmt.Errorf("failed to ensure notes directory: %w", err)
	}

	// Generate filename
	filename, err := generateFilename(notesDir)
	if err != nil {
		return fmt.Errorf("failed to generate filename: %w", err)
	}

	filepath := filepath.Join(notesDir, filename)
	now := time.Now()

	// Create note with empty frontmatter
	note := &Note{
		Filename: filename,
		Frontmatter: Frontmatter{
			Created: NoteTime{now},
			Tags:    []string{},
			Related: []string{},
		},
	}

	if len(args) > 0 {
		// Content provided as argument
		note.Content = "\n" + strings.Join(args, " ") + "\n"
		if err := note.Save(filepath); err != nil {
			return fmt.Errorf("failed to save note: %w", err)
		}
	} else {
		// Open editor
		note.Content = "\n"
		if err := note.Save(filepath); err != nil {
			return fmt.Errorf("failed to save template: %w", err)
		}

		// Open editor
		editor := GetEditor()
		cmd := exec.Command(editor, filepath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			// Clean up file on editor error
			os.Remove(filepath)
			return fmt.Errorf("editor failed: %w", err)
		}

		// Re-read the file to check if content was added
		editedNote, err := ParseNote(filepath)
		if err != nil {
			os.Remove(filepath)
			return fmt.Errorf("failed to parse edited note: %w", err)
		}

		// Check if content is empty or just whitespace
		if strings.TrimSpace(editedNote.Content) == "" {
			os.Remove(filepath)
			fmt.Fprintln(os.Stderr, "Aborted: no content added")
			return nil
		}
	}

	fmt.Printf("Created %s\n", filepath)
	return nil
}

// generateFilename creates a unique filename for the current time
func generateFilename(notesDir string) (string, error) {
	now := time.Now()
	base := now.Format("2006-01-02-1504")

	// Try without suffix first
	filename := base + ".md"
	fullPath := filepath.Join(notesDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return filename, nil
	}

	// Try with suffix
	for i := 1; i < 100; i++ {
		filename = fmt.Sprintf("%s-%d.md", base, i)
		fullPath = filepath.Join(notesDir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return filename, nil
		}
	}

	return "", fmt.Errorf("too many notes in the same minute")
}
