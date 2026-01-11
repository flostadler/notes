# Notes

A minimal, ADHD-friendly notes system written in Go. Designed for quick capture and AI-assisted organization.

## Features

- **Quick Capture**: Create notes instantly from the command line or editor
- **YAML Frontmatter**: Structured metadata (tags, summary, relations)
- **AI Integration**: Generate enrichment prompts for AI assistants
- **Relationship Graphs**: Visualize connections between notes
- **Content Hashing**: Track changes for incremental enrichment

## Installation

```bash
go install notes/cmd/notes@latest
```

Or build from source:

```bash
git clone <repo-url>
cd notes
go build -o notes ./cmd/notes
```

## Project Structure

```
notes/
├── cmd/
│   └── notes/
│       └── main.go         # CLI entry point
├── internal/
│   └── notes/
│       ├── config.go       # Configuration and environment
│       ├── note.go         # Note parsing and rendering
│       ├── meta.go         # Metadata file management
│       ├── cmd_new.go      # Create new notes
│       ├── cmd_list.go     # List notes with filters
│       ├── cmd_show.go     # Display note content
│       ├── cmd_edit.go     # Edit notes in editor
│       ├── cmd_meta.go     # Show note metadata
│       ├── cmd_diff.go     # Find notes needing enrichment
│       ├── cmd_enrich.go   # Generate AI enrichment prompts
│       ├── cmd_update.go   # Update note metadata
│       ├── cmd_sync.go     # Sync metadata from frontmatter
│       ├── cmd_graph.go    # Show relationship graphs
│       ├── cmd_tags.go     # List tags with counts
│       └── *_test.go       # Tests
├── go.mod
├── go.sum
└── README.md
```

## Usage

### Creating Notes

```bash
# Create note with content directly
notes new "Quick thought about project architecture"

# Open editor for new note
notes new
```

### Listing Notes

```bash
# List all notes (newest first)
notes list

# Filter by tags
notes list --tags neo,eval

# Filter by date
notes list --since 2025-01-01

# Limit results
notes list --limit 10

# Show only filenames
notes list --raw
```

### Viewing and Editing

```bash
# Show note content (without frontmatter)
notes show 2025-01-11-1423.md

# Edit note in $EDITOR
notes edit 2025-01-11-1423.md

# Show note metadata as JSON
notes meta 2025-01-11-1423.md
```

### AI-Assisted Enrichment

The enrichment workflow helps you organize notes using AI:

```bash
# Find notes that need enrichment
notes diff

# Generate enrichment prompt for AI
notes enrich

# Update note with AI-generated metadata
notes update 2025-01-11-1423.md \
  --tags "neo,architecture,idea" \
  --summary "Architecture proposal for new service" \
  --related "2025-01-10-0930.md,2025-01-08-1445.md"
```

### Relationship Graphs

```bash
# Show all note connections
notes graph

# Show specific note's neighborhood
notes graph 2025-01-11-1423.md

# Control traversal depth
notes graph 2025-01-11-1423.md --depth 3

# Output as JSON
notes graph --json
```

### Tags

```bash
# List all tags with counts
notes tags
```

### Sync

```bash
# Rebuild .meta.json from frontmatter
notes sync

# Preview changes without writing
notes sync --dry-run

# Force rebuild from scratch
notes sync --force
```

## Note Format

Notes use YAML frontmatter:

```markdown
---
created: 2025-01-11 14:23
tags: [neo, architecture, idea]
summary: "Architecture proposal for new service"
related: [2025-01-10-0930.md]
---

Your note content here...
```

## Metadata

The `.meta.json` file tracks:

- Content hash (SHA256, first 12 chars)
- Enrichment timestamp
- Tags, summary, and relations

This enables incremental enrichment - only notes with changed content need re-processing.

## Environment Variables

| Variable    | Description                    | Default     |
|-------------|--------------------------------|-------------|
| `NOTES_DIR` | Directory for notes            | `~/notes`   |
| `EDITOR`    | Editor for new/edit commands   | `vim`       |

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o notes ./cmd/notes
```

## License

MIT
