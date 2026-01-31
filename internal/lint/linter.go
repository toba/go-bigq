// Package lint provides the core SQL linting logic.
package lint

import (
	"fmt"
	"os"
	"strings"

	"github.com/pacer/go-bigq/bigq"
)

// Result represents a single lint finding.
type Result struct {
	File    string `json:"file"`
	Line    int    `json:"line"`    // 1-based
	Column  int    `json:"column"`  // 1-based
	Level   string `json:"level"`   // "error" or "warning"
	Message string `json:"message"`
}

func (r Result) String() string {
	if r.File != "" && r.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s: %s", r.File, r.Line, r.Column, r.Level, r.Message)
	}
	if r.File != "" {
		return fmt.Sprintf("%s: %s: %s", r.File, r.Level, r.Message)
	}
	return fmt.Sprintf("%s: %s", r.Level, r.Message)
}

// Linter validates SQL statements against a catalog.
type Linter struct {
	catalog *bigq.Catalog
}

// New creates a new Linter with the given catalog.
func New(catalog *bigq.Catalog) *Linter {
	return &Linter{catalog: catalog}
}

// LintSQL checks a SQL string (potentially multi-statement) for errors.
// It uses ZetaSQL's ParseScript to validate the full script including
// scripting constructs (DECLARE, SET, IF, ASSERT, etc.). When a catalog
// is provided, individual non-scripting statements are additionally
// analyzed for schema conformance.
func (l *Linter) LintSQL(sql string) []Result {
	// ParseScript validates the entire script including scripting syntax.
	if err := bigq.ParseScript(sql); err != nil {
		return []Result{{
			Line:    1,
			Column:  1,
			Level:   "error",
			Message: err.Error(),
		}}
	}

	// Without a catalog, syntax validation is all we can do.
	if l.catalog == nil {
		return nil
	}

	// With a catalog, analyze individual statements for schema conformance.
	// AnalyzeStatement doesn't support scripting constructs, so we split
	// and skip those.
	var results []Result
	for _, stmt := range splitStatements(sql) {
		trimmed := strings.TrimSpace(stmt.text)
		if trimmed == "" || trimmed == ";" {
			continue
		}
		if isScriptingStatement(trimmed) {
			continue
		}

		if err := bigq.AnalyzeStatement(trimmed, l.catalog); err != nil {
			results = append(results, Result{
				Line:    stmt.startLine,
				Column:  1,
				Level:   "error",
				Message: err.Error(),
			})
		}
	}
	return results
}

// LintFile reads and lints a SQL file.
func (l *Linter) LintFile(path string) ([]Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	results := l.LintSQL(string(data))
	for i := range results {
		results[i].File = path
	}
	return results, nil
}

// scriptingKeywords are BigQuery scripting keywords that ZetaSQL's
// statement-level analyzer doesn't support. These are skipped during
// schema analysis but are validated by ParseScript.
var scriptingKeywords = []string{
	"DECLARE", "SET", "ASSERT",
	"IF", "ELSEIF", "ELSE", "END",
}

// isScriptingStatement reports whether trimmed starts with a BigQuery
// scripting keyword that ZetaSQL's analyzer cannot handle.
func isScriptingStatement(trimmed string) bool {
	upper := strings.ToUpper(trimmed)
	for _, kw := range scriptingKeywords {
		if !strings.HasPrefix(upper, kw) {
			continue
		}
		// Keyword must be the entire statement or followed by whitespace.
		if len(upper) == len(kw) {
			return true
		}
		switch upper[len(kw)] {
		case ' ', '\t', '\n', '\r':
			return true
		}
	}
	return false
}

type stmtSpan struct {
	text      string
	startLine int
}

// splitStatements splits SQL on semicolons, tracking line numbers.
// Used for per-statement schema analysis when a catalog is provided.
func splitStatements(sql string) []stmtSpan {
	var spans []stmtSpan
	line := 1
	start := 0
	startLine := 1
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(sql); i++ {
		c := sql[i]

		if c == '\n' {
			line++
			if inLineComment {
				inLineComment = false
			}
			continue
		}

		if inLineComment {
			continue
		}

		if inBlockComment {
			if c == '*' && i+1 < len(sql) && sql[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if inSingleQuote {
			if c == '\'' {
				inSingleQuote = false
			} else if c == '\\' {
				i++ // skip escaped char
			}
			continue
		}

		if inDoubleQuote {
			if c == '"' {
				inDoubleQuote = false
			} else if c == '\\' {
				i++
			}
			continue
		}

		if inBacktick {
			if c == '`' {
				inBacktick = false
			}
			continue
		}

		switch c {
		case '\'':
			inSingleQuote = true
		case '"':
			inDoubleQuote = true
		case '`':
			inBacktick = true
		case '-':
			if i+1 < len(sql) && sql[i+1] == '-' {
				inLineComment = true
				i++
			}
		case '/':
			if i+1 < len(sql) && sql[i+1] == '*' {
				inBlockComment = true
				i++
			}
		case ';':
			spans = append(spans, stmtSpan{
				text:      sql[start:i],
				startLine: startLine,
			})
			start = i + 1
			startLine = line
		}
	}

	// Remaining text after last semicolon
	if start < len(sql) {
		remaining := strings.TrimSpace(sql[start:])
		if remaining != "" {
			spans = append(spans, stmtSpan{
				text:      sql[start:],
				startLine: startLine,
			})
		}
	}

	return spans
}
