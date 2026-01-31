#ifndef ZETASQL_BRIDGE_H
#define ZETASQL_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>

// Status result from ZetaSQL operations
typedef struct {
    bool ok;
    char* error_message;      // Caller must free with zetasql_free_string
    int error_line;           // 1-based line, 0 if not available
    int error_column;         // 1-based column, 0 if not available
} zetasql_Status;

// Column info for creating tables
typedef struct {
    const char* name;
    const char* type_name;    // e.g. "INT64", "STRING", "ARRAY<STRING>", "STRUCT<a INT64, b STRING>"
} zetasql_ColumnDef;

// All "new" functions return opaque void* handles.
// All "free" functions take a void* handle.
// All "method" functions take a void* handle as first arg.

// --- TypeFactory ---
void* zetasql_TypeFactory_new();
void zetasql_TypeFactory_free(void* factory);

// --- LanguageOptions ---
void* zetasql_LanguageOptions_new();
void zetasql_LanguageOptions_free(void* opts);
void zetasql_LanguageOptions_EnableMaximumLanguageFeatures(void* opts);
void zetasql_LanguageOptions_SetProductMode(void* opts, int mode);
void zetasql_LanguageOptions_SetSupportsAllStatementKinds(void* opts);

// --- SimpleCatalog ---
void* zetasql_SimpleCatalog_new(const char* name, void* factory);
void zetasql_SimpleCatalog_free(void* catalog);
void zetasql_SimpleCatalog_AddBuiltinFunctionsAndTypes(
    void* catalog, void* lang_opts, zetasql_Status* status);
void* zetasql_SimpleCatalog_AddSubCatalog(void* catalog, const char* name);
void zetasql_SimpleCatalog_AddTable(void* catalog, void* table);

// --- SimpleTable ---
void* zetasql_SimpleTable_new(
    const char* name,
    zetasql_ColumnDef* columns,
    int column_count,
    void* factory,
    zetasql_Status* status);
void zetasql_SimpleTable_free(void* table);

// --- AnalyzerOptions ---
void* zetasql_AnalyzerOptions_new();
void zetasql_AnalyzerOptions_free(void* opts);
void zetasql_AnalyzerOptions_SetLanguageOptions(void* opts, void* lang_opts);

// --- Parse ---
void zetasql_ParseStatement(const char* sql, zetasql_Status* status);
void zetasql_ParseScript(const char* sql, zetasql_Status* status);

// --- Analyze ---
void zetasql_AnalyzeStatement(
    const char* sql, void* catalog, void* opts, zetasql_Status* status);

// --- Utility ---
void zetasql_free_string(char* s);

#ifdef __cplusplus
}
#endif

#endif // ZETASQL_BRIDGE_H
