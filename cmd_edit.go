package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CmdEdit implements the 'notes edit <filename>' command
// Opens note in $EDITOR
func CmdEdit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: notes edit <filename>")
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

	editor := GetEditor()
	cmd := exec.Command(editor, notePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	return nil
}
