package notes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// FileMeta represents metadata for a single note in .meta.json
type FileMeta struct {
	ContentHash string    `json:"content_hash"`
	EnrichedAt  time.Time `json:"enriched_at,omitempty"`
	Tags        []string  `json:"tags"`
	Summary     string    `json:"summary"`
	Related     []string  `json:"related"`
}

// MetaFile represents the .meta.json file structure
type MetaFile struct {
	Files map[string]*FileMeta `json:"files"`
}

// LoadMetaFile loads .meta.json from the notes directory
func LoadMetaFile(notesDir string) (*MetaFile, error) {
	metaPath := filepath.Join(notesDir, ".meta.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty meta file
			return &MetaFile{
				Files: make(map[string]*FileMeta),
			}, nil
		}
		return nil, err
	}

	var meta MetaFile
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	if meta.Files == nil {
		meta.Files = make(map[string]*FileMeta)
	}

	return &meta, nil
}

// Save writes the meta file to disk
func (m *MetaFile) Save(notesDir string) error {
	metaPath := filepath.Join(notesDir, ".meta.json")

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath, data, 0644)
}

// GetFileMeta returns metadata for a specific file
func (m *MetaFile) GetFileMeta(filename string) *FileMeta {
	return m.Files[filename]
}

// SetFileMeta sets metadata for a specific file
func (m *MetaFile) SetFileMeta(filename string, meta *FileMeta) {
	m.Files[filename] = meta
}

// NeedsEnrichment checks if a note needs enrichment
func (m *MetaFile) NeedsEnrichment(filename, currentHash string) bool {
	meta := m.Files[filename]
	if meta == nil {
		return true
	}
	return meta.ContentHash != currentHash
}

// UpdateFromNote updates the meta file entry from a note
func (m *MetaFile) UpdateFromNote(note *Note) {
	filename := filepath.Base(note.Filename)
	meta := m.Files[filename]
	if meta == nil {
		meta = &FileMeta{}
		m.Files[filename] = meta
	}

	meta.ContentHash = note.ContentHash()
	meta.Tags = note.Frontmatter.Tags
	meta.Summary = note.Frontmatter.Summary
	meta.Related = note.Frontmatter.Related
}

// UpdateFromNoteWithEnrichment updates and marks as enriched
func (m *MetaFile) UpdateFromNoteWithEnrichment(note *Note) {
	m.UpdateFromNote(note)
	filename := filepath.Base(note.Filename)
	m.Files[filename].EnrichedAt = time.Now()
}

// AddRelation adds a bidirectional relation between two notes
func (m *MetaFile) AddRelation(from, to string) {
	// Add to -> from relation
	if meta := m.Files[from]; meta != nil {
		if !Contains(meta.Related, to) {
			meta.Related = append(meta.Related, to)
		}
	}

	// Add from -> to relation (bidirectional)
	if meta := m.Files[to]; meta != nil {
		if !Contains(meta.Related, from) {
			meta.Related = append(meta.Related, from)
		}
	}
}

// RemoveRelation removes a bidirectional relation between two notes
func (m *MetaFile) RemoveRelation(from, to string) {
	if meta := m.Files[from]; meta != nil {
		meta.Related = RemoveString(meta.Related, to)
	}
	if meta := m.Files[to]; meta != nil {
		meta.Related = RemoveString(meta.Related, from)
	}
}

// Contains checks if a string slice contains an item
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveString removes an item from a string slice
func RemoveString(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
