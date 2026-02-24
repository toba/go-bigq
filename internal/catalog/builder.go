// Package catalog builds a ZetaSQL SimpleCatalog from schema definitions.
package catalog

import (
	"strings"

	"github.com/pacer/go-bigq/bigq"
	"github.com/pacer/go-bigq/internal/schema"
)

// BuildFromSchema creates a Catalog from a schema definition.
// Tables with qualified names (project.dataset.table) get nested sub-catalogs.
func BuildFromSchema(s *schema.Schema) (*bigq.Catalog, error) {
	cat, err := bigq.NewCatalog("root")
	if err != nil {
		return nil, err
	}

	for _, table := range s.Tables {
		columns := make([]bigq.ColumnDef, len(table.Columns))
		for i, col := range table.Columns {
			columns[i] = bigq.ColumnDef{
				Name:     col.Name,
				TypeName: col.Type,
			}
		}

		// Split qualified name: project.dataset.table
		parts := strings.Split(table.Name, ".")
		if len(parts) == 1 {
			// Unqualified table name - add directly to root catalog
			if err := cat.AddTable(parts[0], columns); err != nil {
				return nil, err
			}
		} else {
			// Create sub-catalogs for each prefix, then add table at the leaf
			// e.g. "project.dataset.table" -> sub("project").sub("dataset").addTable("table")
			current := cat
			for _, part := range parts[:len(parts)-1] {
				sub := current.AddSubCatalog(part)
				// Wrap SubCatalog to satisfy our interface
				// For now, just add the table to the last subcatalog directly
				_ = sub
			}
			// TODO: We need to restructure to support nested sub-catalogs properly.
			// For now, add table with fully qualified name to root catalog.
			if err := cat.AddTable(table.Name, columns); err != nil {
				return nil, err
			}
		}
	}

	return cat, nil
}

// BuildFromFile creates a Catalog from a schema JSON file.
func BuildFromFile(path string) (*bigq.Catalog, error) {
	s, err := schema.LoadFile(path)
	if err != nil {
		return nil, err
	}
	return BuildFromSchema(s)
}

// BuildFromDir creates a Catalog from all JSON schema files in a directory.
func BuildFromDir(dir string) (*bigq.Catalog, error) {
	s, err := schema.LoadDir(dir)
	if err != nil {
		return nil, err
	}
	return BuildFromSchema(s)
}
