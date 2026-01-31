# go-bigq

BigQuery SQL linter. Validates syntax and schema conformance using Google's [ZetaSQL](https://github.com/google/zetasql) parser via CGO.

## Install

```bash
brew tap toba/go-bigq
brew install go-bigq
```

Or build from source:

```bash
make go-bigq
```

## Usage

```bash
# Lint SQL files
go-bigq lint query.sql

# Lint with schema validation
go-bigq lint --schema schema.json query.sql
go-bigq lint --schema-dir schemas/ query.sql

# Read from stdin
echo "SELECT * FORM t" | go-bigq lint --stdin

# Output formats: text (default), json, github-actions
go-bigq lint --format json query.sql
go-bigq lint --format github-actions query.sql
```

Exit code 0 if no errors, 1 if lint errors found, 2 on usage/input errors.

### BigQuery scripting support

go-bigq uses ZetaSQL's `ParseScript` API to natively validate BigQuery scripting syntax — `DECLARE`, `SET`, `ASSERT`, `IF`/`ELSEIF`/`ELSE`/`END IF`, and other procedural constructs are fully parsed and validated alongside your DML/DDL/DQL. No preprocessing or stripping required.

### Schema files

Schema JSON files define table structures for semantic validation:

```json
{
  "project.dataset.table_name": {
    "columns": [
      {"name": "id", "type": "INT64"},
      {"name": "email", "type": "STRING"}
    ]
  }
}
```

Use `--schema` for a single file or `--schema-dir` to load all JSON files from a directory.

### GitHub Actions

```yaml
- name: Lint SQL
  run: go-bigq lint --format github-actions queries/*.sql
```

The `github-actions` format produces `::error` annotations that show inline in pull requests.

## Acknowledgments

go-bigq is built on top of:

- **[Google ZetaSQL](https://github.com/google/zetasql)** (googlesql) — SQL parser and analyzer engine for BigQuery-compatible SQL. Compiled and statically linked via CGO.
- **[ICU](https://icu.unicode.org/)** — International Components for Unicode, linked as a transitive dependency of ZetaSQL.

## License

Apache-2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE) for third-party attribution.
