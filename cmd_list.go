package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CmdList implements the 'notes list' command
func CmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	tagsFlag := fs.String("tags", "", "filter by tags (comma-separated)")
	sinceFlag := fs.String("since", "", "filter by date (YYYY-MM-DD)")
	limitFlag := fs.Int("limit", 20, "limit results")
	rawFlag := fs.Bool("raw", false, "show only filenames")

	if err := fs.Parse(args); err != nil {
		return err
	}

	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	// Parse filters
	var filterTags []string
	if *tagsFlag != "" {
		filterTags = strings.Split(*tagsFlag, ",")
		for i := range filterTags {
			filterTags[i] = strings.TrimSpace(filterTags[i])
		}
	}

	var sinceDate time.Time
	if *sinceFlag != "" {
		var err error
		sinceDate, err = time.Parse("2006-01-02", *sinceFlag)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
	}

	// Find all .md files
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return fmt.Errorf("failed to read notes directory: %w", err)
	}

	type noteInfo struct {
		filename string
		summary  string
		created  time.Time
		tags     []string
	}

	var notes []noteInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		notePath := filepath.Join(notesDir, entry.Name())
		note, err := ParseNote(notePath)
		if err != nil {
			continue
		}

		// Apply date filter
		if !sinceDate.IsZero() && note.Frontmatter.Created.Before(sinceDate) {
			continue
		}

		// Apply tag filter
		if len(filterTags) > 0 && !hasAnyTag(note.Frontmatter.Tags, filterTags) {
			continue
		}

		notes = append(notes, noteInfo{
			filename: entry.Name(),
			summary:  note.GetSummaryOrFirstLine(),
			created:  note.Frontmatter.Created.Time,
			tags:     note.Frontmatter.Tags,
		})
	}

	// Sort by created date, newest first
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].created.After(notes[j].created)
	})

	// Apply limit
	if *limitFlag > 0 && len(notes) > *limitFlag {
		notes = notes[:*limitFlag]
	}

	// Output
	for _, n := range notes {
		if *rawFlag {
			fmt.Println(n.filename)
		} else {
			fmt.Printf("%s  %q\n", n.filename, n.summary)
		}
	}

	return nil
}

func hasAnyTag(noteTags, filterTags []string) bool {
	for _, ft := range filterTags {
		for _, nt := range noteTags {
			if strings.EqualFold(ft, nt) {
				return true
			}
		}
	}
	return false
}
