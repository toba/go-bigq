package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	err := os.WriteFile(path, []byte(`{
		"tables": [{
			"name": "test_table",
			"columns": [
				{"name": "id", "type": "INT64"},
				{"name": "name", "type": "STRING"}
			]
		}]
	}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	s, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if len(s.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(s.Tables))
	}
	if s.Tables[0].Name != "test_table" {
		t.Errorf("table name = %q, want %q", s.Tables[0].Name, "test_table")
	}
	if len(s.Tables[0].Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(s.Tables[0].Columns))
	}
	if s.Tables[0].Columns[0].Type != "INT64" {
		t.Errorf("column type = %q, want %q", s.Tables[0].Columns[0].Type, "INT64")
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()

	for i, data := range []string{
		`{"tables": [{"name": "t1", "columns": [{"name": "a", "type": "INT64"}]}]}`,
		`{"tables": [{"name": "t2", "columns": [{"name": "b", "type": "STRING"}]}]}`,
	} {
		path := filepath.Join(dir, "schema"+string(rune('0'+i))+".json")
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}

	s, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	if len(s.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(s.Tables))
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
