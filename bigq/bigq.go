// Package bigq provides a Go interface to Google's ZetaSQL (googlesql)
// for parsing and analyzing BigQuery SQL.
package bigq

import (
	"github.com/pacer/go-bigq/internal/bridge"
)

// ParseStatement parses a single SQL statement and returns a syntax error if any.
func ParseStatement(sql string) error {
	return bridge.ParseStatement(sql)
}

// ParseScript parses a SQL script (potentially multi-statement, including
// scripting constructs like DECLARE, SET, IF/END IF, ASSERT, etc.) and
// returns a syntax error if any. This is a superset of ParseStatement.
func ParseScript(sql string) error {
	return bridge.ParseScript(sql)
}

// AnalyzeStatement analyzes a SQL statement against a catalog, returning
// an error if the SQL references unknown tables, columns, or functions.
func AnalyzeStatement(sql string, catalog *Catalog) error {
	return bridge.AnalyzeStatement(sql, catalog.inner, catalog.opts)
}

// Catalog holds schema information (tables, functions) used during SQL analysis.
type Catalog struct {
	inner   *bridge.SimpleCatalog
	factory *bridge.TypeFactory
	langOpts *bridge.LanguageOptions
	opts    *bridge.AnalyzerOptions
}

// CatalogOption configures catalog creation.
type CatalogOption func(*catalogConfig)

type catalogConfig struct {
	productMode int
}

// WithProductMode sets the SQL product mode.
// Use bridge.ProductModeExternal (1) for BigQuery compatibility.
func WithProductMode(mode int) CatalogOption {
	return func(c *catalogConfig) {
		c.productMode = mode
	}
}

// NewCatalog creates a new catalog with builtin BigQuery functions and types.
func NewCatalog(name string, options ...CatalogOption) (*Catalog, error) {
	cfg := &catalogConfig{
		productMode: bridge.ProductModeExternal, // BigQuery mode by default
	}
	for _, opt := range options {
		opt(cfg)
	}

	factory := bridge.NewTypeFactory()
	langOpts := bridge.NewLanguageOptions()
	langOpts.EnableMaximumLanguageFeatures()
	langOpts.SetProductMode(cfg.productMode)
	langOpts.SetSupportsAllStatementKinds()

	catalog := bridge.NewSimpleCatalog(name, factory)
	if err := catalog.AddBuiltinFunctionsAndTypes(langOpts); err != nil {
		return nil, err
	}

	analyzerOpts := bridge.NewAnalyzerOptions()
	analyzerOpts.SetLanguageOptions(langOpts)

	return &Catalog{
		inner:    catalog,
		factory:  factory,
		langOpts: langOpts,
		opts:     analyzerOpts,
	}, nil
}

// AddTable adds a table to the catalog.
// The table name can be qualified (e.g. "project.dataset.table") - intermediate
// sub-catalogs will be created automatically.
func (c *Catalog) AddTable(name string, columns []ColumnDef) error {
	return c.inner.AddTable(name, toBridgeColumns(columns))
}

// AddSubCatalog adds a named sub-catalog (e.g. for a dataset).
func (c *Catalog) AddSubCatalog(name string) *SubCatalog {
	sub := c.inner.AddSubCatalog(name)
	return &SubCatalog{inner: sub}
}

// Close releases all resources held by the catalog.
func (c *Catalog) Close() {
	if c.inner != nil {
		c.inner.Close()
	}
	if c.opts != nil {
		c.opts.Close()
	}
	if c.langOpts != nil {
		c.langOpts.Close()
	}
	if c.factory != nil {
		c.factory.Close()
	}
}

// SubCatalog represents a nested catalog (e.g. a dataset).
type SubCatalog struct {
	inner *bridge.SimpleCatalog
}

// AddTable adds a table to this sub-catalog.
func (s *SubCatalog) AddTable(name string, columns []ColumnDef) error {
	return s.inner.AddTable(name, toBridgeColumns(columns))
}

// ColumnDef defines a table column.
type ColumnDef struct {
	Name     string
	TypeName string // BigQuery type: INT64, STRING, ARRAY<STRING>, STRUCT<a INT64, b STRING>, etc.
}

func toBridgeColumns(columns []ColumnDef) []bridge.ColumnDef {
	out := make([]bridge.ColumnDef, len(columns))
	for i, c := range columns {
		out[i] = bridge.ColumnDef{Name: c.Name, TypeName: c.TypeName}
	}
	return out
}
