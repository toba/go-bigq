package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pacer/go-bigq/bigq"
	"github.com/pacer/go-bigq/internal/catalog"
	"github.com/pacer/go-bigq/internal/lint"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Usage: go-bigq lint [flags] [files...]")
		return 2
	}

	switch args[0] {
	case "lint":
		return runLint(args[1:], stdout, stderr)
	case "version":
		fmt.Fprintln(stdout, "go-bigq version "+version)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n", args[0])
		fmt.Fprintln(stderr, "Usage: go-bigq lint [flags] [files...]")
		return 2
	}
}

func runLint(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	fs.SetOutput(stderr)

	schemaPath := fs.String("schema", "", "Path to schema JSON file")
	schemaDir := fs.String("schema-dir", "", "Directory of schema JSON files")
	format := fs.String("format", "text", "Output format: text, json, github-actions")
	useStdin := fs.Bool("stdin", false, "Read SQL from stdin")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Build catalog from schema
	var cat *bigq.Catalog
	if *schemaPath != "" {
		var err error
		cat, err = catalog.BuildFromFile(*schemaPath)
		if err != nil {
			fmt.Fprintf(stderr, "Error loading schema: %s\n", err)
			return 2
		}
		defer cat.Close()
	} else if *schemaDir != "" {
		var err error
		cat, err = catalog.BuildFromDir(*schemaDir)
		if err != nil {
			fmt.Fprintf(stderr, "Error loading schema directory: %s\n", err)
			return 2
		}
		defer cat.Close()
	}

	linter := lint.New(cat)
	var allResults []lint.Result

	if *useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading stdin: %s\n", err)
			return 2
		}
		results := linter.LintSQL(string(data))
		for i := range results {
			results[i].File = "<stdin>"
		}
		allResults = append(allResults, results...)
	}

	files := fs.Args()
	for _, file := range files {
		results, err := linter.LintFile(file)
		if err != nil {
			fmt.Fprintf(stderr, "Error: %s\n", err)
			return 2
		}
		allResults = append(allResults, results...)
	}

	if len(files) == 0 && !*useStdin {
		fmt.Fprintln(stderr, "No input files. Use --stdin or pass file paths.")
		return 2
	}

	// Output results
	switch *format {
	case "json":
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		enc.Encode(allResults)
	case "github-actions":
		for _, r := range allResults {
			fmt.Fprintf(stdout, "::error file=%s,line=%d,col=%d::%s\n",
				r.File, r.Line, r.Column, r.Message)
		}
	default: // text
		for _, r := range allResults {
			fmt.Fprintln(stdout, r.String())
		}
	}

	if len(allResults) > 0 {
		return 1
	}
	return 0
}
