// Package schema handles loading table schemas from JSON files.
package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Schema represents a collection of table definitions.
type Schema struct {
	Tables []Table `json:"tables"`
}

// Table represents a table definition.
type Table struct {
	Name    string   `json:"name"` // Fully qualified: project.dataset.table
	Columns []Column `json:"columns"`
}

// Column represents a column definition.
type Column struct {
	Name string `json:"name"`
	Type string `json:"type"` // BigQuery type: INT64, STRING, ARRAY<STRING>, etc.
}

// LoadFile loads a schema from a JSON file.
func LoadFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading schema file %s: %w", path, err)
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing schema file %s: %w", path, err)
	}
	return &s, nil
}

// LoadDir loads all .json schema files from a directory.
func LoadDir(dir string) (*Schema, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading schema directory %s: %w", dir, err)
	}

	merged := &Schema{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		s, err := LoadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		merged.Tables = append(merged.Tables, s.Tables...)
	}
	return merged, nil
}
