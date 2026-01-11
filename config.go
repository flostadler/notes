package main

import (
	"os"
	"path/filepath"
)

// GetNotesDir returns the notes directory path
// Uses NOTES_DIR env var if set, otherwise defaults to ~/notes
func GetNotesDir() (string, error) {
	if dir := os.Getenv("NOTES_DIR"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "notes"), nil
}

// EnsureNotesDir creates the notes directory if it doesn't exist
func EnsureNotesDir() (string, error) {
	dir, err := GetNotesDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}

// GetEditor returns the editor to use
func GetEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vim"
}

// NormalizeFilename ensures a filename has .md extension
func NormalizeFilename(filename string) string {
	if filepath.Ext(filename) != ".md" {
		return filename + ".md"
	}
	return filename
}
