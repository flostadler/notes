package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CmdGraph implements the 'notes graph [filename]' command
func CmdGraph(args []string) error {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	depthFlag := fs.Int("depth", 2, "how many hops to traverse")
	jsonFlag := fs.Bool("json", false, "output as JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}

	notesDir, err := GetNotesDir()
	if err != nil {
		return fmt.Errorf("failed to get notes directory: %w", err)
	}

	meta, err := LoadMetaFile(notesDir)
	if err != nil {
		return fmt.Errorf("failed to load meta file: %w", err)
	}

	remaining := fs.Args()

	if len(remaining) > 0 {
		// Show specific note's neighborhood
		filename := NormalizeFilename(remaining[0])
		return showNeighborhood(notesDir, meta, filename, *depthFlag, *jsonFlag)
	}

	// Show all connections
	return showAllConnections(meta, *jsonFlag)
}

func showAllConnections(meta *MetaFile, asJSON bool) error {
	if asJSON {
		type connection struct {
			From       string   `json:"from"`
			To         []string `json:"to"`
			SharedTags []string `json:"shared_tags,omitempty"`
		}
		var connections []connection
		for filename, fileMeta := range meta.Files {
			if len(fileMeta.Related) > 0 {
				conn := connection{
					From: filename,
					To:   fileMeta.Related,
				}
				connections = append(connections, conn)
			}
		}
		data, err := json.MarshalIndent(connections, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Sort filenames for consistent output
	var filenames []string
	for filename := range meta.Files {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		fileMeta := meta.Files[filename]
		if len(fileMeta.Related) == 0 {
			continue
		}

		fmt.Println(filename)
		for _, rel := range fileMeta.Related {
			sharedTags := getSharedTags(meta, filename, rel)
			if len(sharedTags) > 0 {
				fmt.Printf("  → %s (%s)\n", rel, strings.Join(sharedTags, ", "))
			} else {
				fmt.Printf("  → %s\n", rel)
			}
		}
	}

	return nil
}

func showNeighborhood(notesDir string, meta *MetaFile, filename string, depth int, asJSON bool) error {
	// Verify file exists
	notePath := filepath.Join(notesDir, filename)
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("note not found: %s", filename)
	}

	// Get summary for the root note
	rootSummary := getSummary(notesDir, meta, filename)

	if asJSON {
		type graphNode struct {
			Filename string      `json:"filename"`
			Summary  string      `json:"summary,omitempty"`
			Related  []graphNode `json:"related,omitempty"`
		}

		visited := make(map[string]bool)
		var buildGraph func(f string, d int) graphNode
		buildGraph = func(f string, d int) graphNode {
			node := graphNode{
				Filename: f,
				Summary:  getSummary(notesDir, meta, f),
			}
			if d <= 0 || visited[f] {
				return node
			}
			visited[f] = true

			if fileMeta := meta.GetFileMeta(f); fileMeta != nil {
				for _, rel := range fileMeta.Related {
					node.Related = append(node.Related, buildGraph(rel, d-1))
				}
			}
			return node
		}

		root := buildGraph(filename, depth)
		data, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Text output with tree structure
	fmt.Printf("%s %q\n", filename, rootSummary)

	visited := make(map[string]bool)
	visited[filename] = true

	fileMeta := meta.GetFileMeta(filename)
	if fileMeta == nil {
		return nil
	}

	printTree(notesDir, meta, fileMeta.Related, depth-1, "", visited)
	return nil
}

func printTree(notesDir string, meta *MetaFile, related []string, depth int, prefix string, visited map[string]bool) {
	for i, rel := range related {
		isLast := i == len(related)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		summary := getSummary(notesDir, meta, rel)
		fmt.Printf("%s%s%s %q\n", prefix, connector, rel, summary)

		if depth > 0 && !visited[rel] {
			visited[rel] = true
			if fileMeta := meta.GetFileMeta(rel); fileMeta != nil && len(fileMeta.Related) > 0 {
				// Filter out already visited nodes
				var unvisited []string
				for _, r := range fileMeta.Related {
					if !visited[r] {
						unvisited = append(unvisited, r)
					}
				}
				if len(unvisited) > 0 {
					printTree(notesDir, meta, unvisited, depth-1, childPrefix, visited)
				}
			}
		}
	}
}

func getSummary(notesDir string, meta *MetaFile, filename string) string {
	if fileMeta := meta.GetFileMeta(filename); fileMeta != nil && fileMeta.Summary != "" {
		return fileMeta.Summary
	}

	// Try to get from file
	notePath := filepath.Join(notesDir, filename)
	if note, err := ParseNote(notePath); err == nil {
		return note.GetSummaryOrFirstLine()
	}

	return ""
}

func getSharedTags(meta *MetaFile, file1, file2 string) []string {
	meta1 := meta.GetFileMeta(file1)
	meta2 := meta.GetFileMeta(file2)

	if meta1 == nil || meta2 == nil {
		return nil
	}

	var shared []string
	for _, t1 := range meta1.Tags {
		for _, t2 := range meta2.Tags {
			if strings.EqualFold(t1, t2) {
				shared = append(shared, t1)
				break
			}
		}
	}
	return shared
}
