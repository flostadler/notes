package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CmdTags implements the 'notes tags' command
// Lists all tags with counts
func CmdTags(args []string) error {
	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	// Collect tags from all notes
	tagCounts := make(map[string]int)

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
			continue
		}

		for _, tag := range note.Frontmatter.Tags {
			tagCounts[strings.ToLower(tag)]++
		}
	}

	if len(tagCounts) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	// Sort by count (descending), then alphabetically
	type tagCount struct {
		tag   string
		count int
	}

	var tags []tagCount
	for tag, count := range tagCounts {
		tags = append(tags, tagCount{tag, count})
	}

	sort.Slice(tags, func(i, j int) bool {
		if tags[i].count != tags[j].count {
			return tags[i].count > tags[j].count
		}
		return tags[i].tag < tags[j].tag
	})

	for _, tc := range tags {
		fmt.Printf("%s (%d)\n", tc.tag, tc.count)
	}

	return nil
}
