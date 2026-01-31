// Package bridge provides the CGO bridge to ZetaSQL's C++ library.
// It compiles zetasql_bridge.cc and links against the pre-built static library.
package bridge

// #cgo CXXFLAGS: -std=c++20 -I${SRCDIR}/../../lib/include
// #cgo LDFLAGS: -L${SRCDIR}/../../lib -lzetasql -licuuc -licui18n -licudata -lstdc++ -lm -lpthread -lc++
// #include "zetasql_bridge.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

// Status represents the result of a ZetaSQL operation.
type Status struct {
	OK           bool
	ErrorMessage string
	ErrorLine    int // 1-based, 0 if not available
	ErrorColumn  int // 1-based, 0 if not available
}

func statusFromC(s C.zetasql_Status) Status {
	st := Status{
		OK:          bool(s.ok),
		ErrorLine:   int(s.error_line),
		ErrorColumn: int(s.error_column),
	}
	if s.error_message != nil {
		st.ErrorMessage = C.GoString(s.error_message)
		C.zetasql_free_string(s.error_message)
	}
	return st
}

func (s Status) Error() string {
	if s.OK {
		return ""
	}
	if s.ErrorLine > 0 {
		return fmt.Sprintf("%d:%d: %s", s.ErrorLine, s.ErrorColumn, s.ErrorMessage)
	}
	return s.ErrorMessage
}

// TypeFactory manages ZetaSQL type objects.
type TypeFactory struct {
	raw unsafe.Pointer
}

func NewTypeFactory() *TypeFactory {
	tf := &TypeFactory{raw: C.zetasql_TypeFactory_new()}
	runtime.SetFinalizer(tf, func(t *TypeFactory) { t.Close() })
	return tf
}

func (tf *TypeFactory) Close() {
	if tf.raw != nil {
		C.zetasql_TypeFactory_free(tf.raw)
		tf.raw = nil
	}
}

// LanguageOptions controls which SQL features are enabled.
type LanguageOptions struct {
	raw unsafe.Pointer
}

func NewLanguageOptions() *LanguageOptions {
	lo := &LanguageOptions{raw: C.zetasql_LanguageOptions_new()}
	runtime.SetFinalizer(lo, func(l *LanguageOptions) { l.Close() })
	return lo
}

func (lo *LanguageOptions) Close() {
	if lo.raw != nil {
		C.zetasql_LanguageOptions_free(lo.raw)
		lo.raw = nil
	}
}

func (lo *LanguageOptions) EnableMaximumLanguageFeatures() {
	C.zetasql_LanguageOptions_EnableMaximumLanguageFeatures(lo.raw)
}

// ProductMode constants
const (
	ProductModeInternal = 0 // PRODUCT_INTERNAL
	ProductModeExternal = 1 // PRODUCT_EXTERNAL (BigQuery mode)
)

func (lo *LanguageOptions) SetProductMode(mode int) {
	C.zetasql_LanguageOptions_SetProductMode(lo.raw, C.int(mode))
}

func (lo *LanguageOptions) SetSupportsAllStatementKinds() {
	C.zetasql_LanguageOptions_SetSupportsAllStatementKinds(lo.raw)
}

// SimpleCatalog holds schema information for SQL analysis.
type SimpleCatalog struct {
	raw     unsafe.Pointer
	factory *TypeFactory // kept alive for the catalog's lifetime
}

func NewSimpleCatalog(name string, factory *TypeFactory) *SimpleCatalog {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cat := &SimpleCatalog{
		raw:     C.zetasql_SimpleCatalog_new(cname, factory.raw),
		factory: factory,
	}
	runtime.SetFinalizer(cat, func(c *SimpleCatalog) { c.Close() })
	return cat
}

func (c *SimpleCatalog) Close() {
	if c.raw != nil {
		C.zetasql_SimpleCatalog_free(c.raw)
		c.raw = nil
	}
}

func (c *SimpleCatalog) AddBuiltinFunctionsAndTypes(langOpts *LanguageOptions) error {
	var st C.zetasql_Status
	C.zetasql_SimpleCatalog_AddBuiltinFunctionsAndTypes(c.raw, langOpts.raw, &st)
	status := statusFromC(st)
	if !status.OK {
		return fmt.Errorf("AddBuiltinFunctionsAndTypes: %s", status.Error())
	}
	return nil
}

func (c *SimpleCatalog) AddSubCatalog(name string) *SimpleCatalog {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	sub := C.zetasql_SimpleCatalog_AddSubCatalog(c.raw, cname)
	return &SimpleCatalog{raw: sub, factory: c.factory}
}

// ColumnDef defines a column for table creation.
type ColumnDef struct {
	Name     string
	TypeName string // e.g. "INT64", "STRING", "ARRAY<STRING>"
}

func (c *SimpleCatalog) AddTable(name string, columns []ColumnDef) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cColumns := make([]C.zetasql_ColumnDef, len(columns))
	cStrings := make([]*C.char, len(columns)*2)
	for i, col := range columns {
		cn := C.CString(col.Name)
		ct := C.CString(col.TypeName)
		cStrings[i*2] = cn
		cStrings[i*2+1] = ct
		cColumns[i].name = cn
		cColumns[i].type_name = ct
	}
	defer func() {
		for _, s := range cStrings {
			C.free(unsafe.Pointer(s))
		}
	}()

	var colPtr *C.zetasql_ColumnDef
	if len(cColumns) > 0 {
		colPtr = &cColumns[0]
	}

	var st C.zetasql_Status
	tableRaw := C.zetasql_SimpleTable_new(cname, colPtr, C.int(len(columns)), c.factory.raw, &st)
	status := statusFromC(st)
	if !status.OK {
		return fmt.Errorf("create table %s: %s", name, status.Error())
	}

	C.zetasql_SimpleCatalog_AddTable(c.raw, tableRaw)
	return nil
}

// AnalyzerOptions controls analysis behavior.
type AnalyzerOptions struct {
	raw unsafe.Pointer
}

func NewAnalyzerOptions() *AnalyzerOptions {
	ao := &AnalyzerOptions{raw: C.zetasql_AnalyzerOptions_new()}
	runtime.SetFinalizer(ao, func(a *AnalyzerOptions) { a.Close() })
	return ao
}

func (ao *AnalyzerOptions) Close() {
	if ao.raw != nil {
		C.zetasql_AnalyzerOptions_free(ao.raw)
		ao.raw = nil
	}
}

func (ao *AnalyzerOptions) SetLanguageOptions(langOpts *LanguageOptions) {
	C.zetasql_AnalyzerOptions_SetLanguageOptions(ao.raw, langOpts.raw)
}

// ParseStatement parses a SQL statement and returns any syntax error.
func ParseStatement(sql string) error {
	csql := C.CString(sql)
	defer C.free(unsafe.Pointer(csql))

	var st C.zetasql_Status
	C.zetasql_ParseStatement(csql, &st)
	status := statusFromC(st)
	if !status.OK {
		return fmt.Errorf("parse error: %s", status.Error())
	}
	return nil
}

// ParseScript parses a SQL script (potentially multi-statement, with
// scripting constructs like DECLARE, SET, IF, etc.) and returns any syntax error.
func ParseScript(sql string) error {
	csql := C.CString(sql)
	defer C.free(unsafe.Pointer(csql))

	var st C.zetasql_Status
	C.zetasql_ParseScript(csql, &st)
	status := statusFromC(st)
	if !status.OK {
		return fmt.Errorf("parse error: %s", status.Error())
	}
	return nil
}

// AnalyzeStatement analyzes a SQL statement against a catalog.
func AnalyzeStatement(sql string, catalog *SimpleCatalog, opts *AnalyzerOptions) error {
	csql := C.CString(sql)
	defer C.free(unsafe.Pointer(csql))

	var st C.zetasql_Status
	C.zetasql_AnalyzeStatement(csql, catalog.raw, opts.raw, &st)
	status := statusFromC(st)
	if !status.OK {
		return fmt.Errorf("analysis error: %s", status.Error())
	}
	return nil
}
