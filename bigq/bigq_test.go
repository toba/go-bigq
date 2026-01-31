package bigq_test

import (
	"testing"

	"github.com/pacer/go-bigq/bigq"
)

func TestParseStatement(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{"simple select", "SELECT 1", false},
		{"select with alias", "SELECT 1 + 2 AS result", false},
		{"select star", "SELECT * FROM t", false},
		{"syntax error", "SELECT * FORM t", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bigq.ParseStatement(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStatement(%q) error = %v, wantErr %v", tt.sql, err, tt.wantErr)
			}
		})
	}
}

func TestParseScript(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		// Regular statements still work (superset of ParseStatement)
		{"simple select", "SELECT 1;", false},
		{"multi statement", "SELECT 1; SELECT 2;", false},

		// Scripting constructs
		{"declare", "DECLARE x INT64;", false},
		{"declare with default", "DECLARE x INT64 DEFAULT 42;", false},
		{"set", "DECLARE x INT64; SET x = 1;", false},
		{"assert", "ASSERT 1 > 0;", false},
		{"assert with message", "ASSERT 1 > 0 AS 'failed';", false},
		{"if block", "IF true THEN SELECT 1; END IF;", false},
		{"if else", "IF true THEN SELECT 1; ELSE SELECT 2; END IF;", false},
		{"if elseif else", "IF true THEN SELECT 1; ELSEIF false THEN SELECT 2; ELSE SELECT 3; END IF;", false},
		{"full script", "DECLARE x INT64 DEFAULT 1; IF x > 0 THEN SELECT x; END IF;", false},

		// Syntax errors
		{"syntax error", "SELECT * FORM t;", true},
		{"bad declare", "DECLARE;", true},
		{"unclosed if", "IF true THEN SELECT 1;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bigq.ParseScript(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseScript(%q) error = %v, wantErr %v", tt.sql, err, tt.wantErr)
			}
		})
	}
}

func TestAnalyzeStatement(t *testing.T) {
	cat, err := bigq.NewCatalog("test")
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	defer cat.Close()

	err = cat.AddTable("my_table", []bigq.ColumnDef{
		{Name: "id", TypeName: "INT64"},
		{Name: "name", TypeName: "STRING"},
		{Name: "created_at", TypeName: "TIMESTAMP"},
	})
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{"valid select", "SELECT id, name FROM my_table", false},
		{"select star", "SELECT * FROM my_table", false},
		{"bad column", "SELECT nonexistent FROM my_table", true},
		{"bad table", "SELECT 1 FROM no_such_table", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bigq.AnalyzeStatement(tt.sql, cat)
			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeStatement(%q) error = %v, wantErr %v", tt.sql, err, tt.wantErr)
			}
		})
	}
}
