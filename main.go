package main

import (
	"fmt"
	"os"
)

const usage = `notes - A minimal, ADHD-friendly notes system

Usage:
  notes <command> [arguments]

Commands:
  new [content]     Create a new note (opens editor if no content provided)
  list              List all notes, newest first
  show <filename>   Print note content (without frontmatter)
  edit <filename>   Open note in $EDITOR
  meta <filename>   Print note metadata as JSON

  diff              List notes that need enrichment
  enrich            Output enrichment prompt for AI
  update <file>     Update note metadata (used by AI)
  sync              Rebuild .meta.json from frontmatter

  graph [filename]  Show relationship graph
  tags              List all tags with counts

Flags vary by command. Use 'notes <command> --help' for details.

Environment:
  NOTES_DIR   Notes directory (default: ~/notes)
  EDITOR      Editor for new/edit (default: vim)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "new":
		err = CmdNew(args)
	case "list":
		err = CmdList(args)
	case "show":
		err = CmdShow(args)
	case "edit":
		err = CmdEdit(args)
	case "meta":
		err = CmdMeta(args)
	case "diff":
		err = CmdDiff(args)
	case "enrich":
		err = CmdEnrich(args)
	case "update":
		err = CmdUpdate(args)
	case "sync":
		err = CmdSync(args)
	case "graph":
		err = CmdGraph(args)
	case "tags":
		err = CmdTags(args)
	case "help", "-h", "--help":
		fmt.Print(usage)
	case "version", "-v", "--version":
		fmt.Println("notes v0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintf(os.Stderr, "Run 'notes help' for usage.\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
