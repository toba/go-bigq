#include "zetasql_bridge.h"

#include <cstdlib>
#include <cstring>
#include <memory>
#include <string>
#include <vector>

#include "zetasql/public/analyzer.h"
#include "zetasql/public/analyzer_options.h"
#include "zetasql/public/catalog.h"
#include "zetasql/public/error_helpers.h"
#include "zetasql/public/language_options.h"
#include "zetasql/public/simple_catalog.h"
#include "zetasql/public/type.h"
#include "zetasql/public/types/type_factory.h"
#include "zetasql/public/builtin_function_options.h"
#include "zetasql/parser/parser.h"
#include "absl/status/status.h"
#include "absl/strings/string_view.h"

static char* dup_string(const std::string& s) {
    char* out = static_cast<char*>(malloc(s.size() + 1));
    memcpy(out, s.data(), s.size());
    out[s.size()] = '\0';
    return out;
}

static void set_status(zetasql_Status* st, const absl::Status& status) {
    if (status.ok()) {
        st->ok = true;
        st->error_message = nullptr;
        st->error_line = 0;
        st->error_column = 0;
    } else {
        st->ok = false;
        st->error_message = dup_string(std::string(status.message()));
        st->error_line = 0;
        st->error_column = 0;

        zetasql::ErrorLocation location;
        if (zetasql::GetErrorLocation(status, &location)) {
            st->error_line = location.line();
            st->error_column = location.column();
        }
    }
}

static absl::Status parse_type(const std::string& type_str,
                                zetasql::TypeFactory* factory,
                                const zetasql::Type** out_type) {
    static const struct { const char* name; zetasql::TypeKind kind; } simple_types[] = {
        {"INT64", zetasql::TYPE_INT64},
        {"INT32", zetasql::TYPE_INT32},
        {"UINT32", zetasql::TYPE_UINT32},
        {"UINT64", zetasql::TYPE_UINT64},
        {"FLOAT32", zetasql::TYPE_FLOAT},
        {"FLOAT64", zetasql::TYPE_DOUBLE},
        {"FLOAT", zetasql::TYPE_FLOAT},
        {"DOUBLE", zetasql::TYPE_DOUBLE},
        {"NUMERIC", zetasql::TYPE_NUMERIC},
        {"BIGNUMERIC", zetasql::TYPE_BIGNUMERIC},
        {"BOOL", zetasql::TYPE_BOOL},
        {"BOOLEAN", zetasql::TYPE_BOOL},
        {"STRING", zetasql::TYPE_STRING},
        {"BYTES", zetasql::TYPE_BYTES},
        {"DATE", zetasql::TYPE_DATE},
        {"DATETIME", zetasql::TYPE_DATETIME},
        {"TIME", zetasql::TYPE_TIME},
        {"TIMESTAMP", zetasql::TYPE_TIMESTAMP},
        {"GEOGRAPHY", zetasql::TYPE_GEOGRAPHY},
        {"JSON", zetasql::TYPE_JSON},
        {"INTERVAL", zetasql::TYPE_INTERVAL},
    };

    std::string upper = type_str;
    for (auto& c : upper) c = toupper(c);

    size_t start = upper.find_first_not_of(" \t\n\r");
    size_t end = upper.find_last_not_of(" \t\n\r");
    if (start == std::string::npos) {
        return absl::InvalidArgumentError("Empty type string");
    }
    upper = upper.substr(start, end - start + 1);
    std::string original_trimmed = type_str.substr(start, end - start + 1);

    for (const auto& st : simple_types) {
        if (upper == st.name) {
            *out_type = zetasql::types::TypeFromSimpleTypeKind(st.kind);
            return absl::OkStatus();
        }
    }

    if (upper.size() > 7 && upper.substr(0, 6) == "ARRAY<" && upper.back() == '>') {
        std::string inner = original_trimmed.substr(6, original_trimmed.size() - 7);
        const zetasql::Type* element_type = nullptr;
        auto s = parse_type(inner, factory, &element_type);
        if (!s.ok()) return s;
        return factory->MakeArrayType(element_type, out_type);
    }

    if (upper.size() > 8 && upper.substr(0, 7) == "STRUCT<" && upper.back() == '>') {
        std::string inner = original_trimmed.substr(7, original_trimmed.size() - 8);
        std::vector<zetasql::StructType::StructField> fields;

        int depth = 0;
        size_t field_start = 0;
        for (size_t i = 0; i <= inner.size(); i++) {
            if (i == inner.size() || (inner[i] == ',' && depth == 0)) {
                std::string field_str = inner.substr(field_start, i - field_start);
                size_t fs = field_str.find_first_not_of(" \t");
                size_t fe = field_str.find_last_not_of(" \t");
                if (fs == std::string::npos) {
                    return absl::InvalidArgumentError("Empty field in STRUCT");
                }
                field_str = field_str.substr(fs, fe - fs + 1);

                size_t split = std::string::npos;
                int d2 = 0;
                for (size_t j = 0; j < field_str.size(); j++) {
                    if (field_str[j] == '<') d2++;
                    else if (field_str[j] == '>') d2--;
                    else if (field_str[j] == ' ' && d2 == 0) {
                        split = j;
                        break;
                    }
                }
                if (split == std::string::npos) {
                    return absl::InvalidArgumentError(
                        "Invalid STRUCT field (expected 'name type'): " + field_str);
                }

                std::string field_name = field_str.substr(0, split);
                std::string field_type_str = field_str.substr(split + 1);
                size_t fts = field_type_str.find_first_not_of(" \t");
                if (fts != std::string::npos) field_type_str = field_type_str.substr(fts);

                const zetasql::Type* field_type = nullptr;
                auto s = parse_type(field_type_str, factory, &field_type);
                if (!s.ok()) return s;
                fields.push_back({field_name, field_type});

                field_start = i + 1;
            } else if (inner[i] == '<') {
                depth++;
            } else if (inner[i] == '>') {
                depth--;
            }
        }

        return factory->MakeStructType(fields, out_type);
    }

    return absl::InvalidArgumentError("Unknown type: " + type_str);
}


extern "C" {

void* zetasql_TypeFactory_new() {
    return static_cast<void*>(new zetasql::TypeFactory());
}

void zetasql_TypeFactory_free(void* factory) {
    delete static_cast<zetasql::TypeFactory*>(factory);
}

void* zetasql_LanguageOptions_new() {
    return static_cast<void*>(new zetasql::LanguageOptions());
}

void zetasql_LanguageOptions_free(void* opts) {
    delete static_cast<zetasql::LanguageOptions*>(opts);
}

void zetasql_LanguageOptions_EnableMaximumLanguageFeatures(void* opts) {
    static_cast<zetasql::LanguageOptions*>(opts)->EnableMaximumLanguageFeatures();
}

void zetasql_LanguageOptions_SetProductMode(void* opts, int mode) {
    static_cast<zetasql::LanguageOptions*>(opts)->set_product_mode(
        static_cast<zetasql::ProductMode>(mode));
}

void zetasql_LanguageOptions_SetSupportsAllStatementKinds(void* opts) {
    static_cast<zetasql::LanguageOptions*>(opts)->SetSupportsAllStatementKinds();
}

void* zetasql_SimpleCatalog_new(const char* name, void* factory) {
    return static_cast<void*>(
        new zetasql::SimpleCatalog(name, static_cast<zetasql::TypeFactory*>(factory)));
}

void zetasql_SimpleCatalog_free(void* catalog) {
    delete static_cast<zetasql::SimpleCatalog*>(catalog);
}

void zetasql_SimpleCatalog_AddBuiltinFunctionsAndTypes(
    void* catalog, void* lang_opts, zetasql_Status* status) {
    zetasql::BuiltinFunctionOptions options(
        *static_cast<zetasql::LanguageOptions*>(lang_opts));
    auto s = static_cast<zetasql::SimpleCatalog*>(catalog)->AddBuiltinFunctionsAndTypes(options);
    set_status(status, s);
}

void* zetasql_SimpleCatalog_AddSubCatalog(void* catalog, const char* name) {
    auto* parent = static_cast<zetasql::SimpleCatalog*>(catalog);
    auto* sub = new zetasql::SimpleCatalog(name, parent->type_factory());
    parent->AddOwnedCatalog(sub);
    return static_cast<void*>(sub);
}

void zetasql_SimpleCatalog_AddTable(void* catalog, void* table) {
    static_cast<zetasql::SimpleCatalog*>(catalog)->AddTable(
        static_cast<zetasql::SimpleTable*>(table));
}

void* zetasql_SimpleTable_new(
    const char* name,
    zetasql_ColumnDef* columns,
    int column_count,
    void* factory,
    zetasql_Status* status) {
    auto* tf = static_cast<zetasql::TypeFactory*>(factory);

    std::vector<zetasql::SimpleTable::NameAndType> cols;
    for (int i = 0; i < column_count; i++) {
        const zetasql::Type* col_type = nullptr;
        auto s = parse_type(columns[i].type_name, tf, &col_type);
        if (!s.ok()) {
            set_status(status, s);
            return nullptr;
        }
        cols.push_back({columns[i].name, col_type});
    }

    auto* table = new zetasql::SimpleTable(name, cols);

    status->ok = true;
    status->error_message = nullptr;
    status->error_line = 0;
    status->error_column = 0;

    return static_cast<void*>(table);
}

void zetasql_SimpleTable_free(void* table) {
    delete static_cast<zetasql::SimpleTable*>(table);
}

void* zetasql_AnalyzerOptions_new() {
    return static_cast<void*>(new zetasql::AnalyzerOptions());
}

void zetasql_AnalyzerOptions_free(void* opts) {
    delete static_cast<zetasql::AnalyzerOptions*>(opts);
}

void zetasql_AnalyzerOptions_SetLanguageOptions(void* opts, void* lang_opts) {
    static_cast<zetasql::AnalyzerOptions*>(opts)->set_language(
        *static_cast<zetasql::LanguageOptions*>(lang_opts));
}

void zetasql_ParseStatement(const char* sql, zetasql_Status* status) {
    zetasql::LanguageOptions lang;
    lang.EnableMaximumLanguageFeatures();
    lang.SetSupportsAllStatementKinds();
    zetasql::ParserOptions opts(lang);
    std::unique_ptr<zetasql::ParserOutput> output;
    auto s = zetasql::ParseStatement(sql, opts, &output);
    set_status(status, s);
}

void zetasql_ParseScript(const char* sql, zetasql_Status* status) {
    zetasql::LanguageOptions lang;
    lang.EnableMaximumLanguageFeatures();
    lang.SetSupportsAllStatementKinds();
    zetasql::ParserOptions opts(lang);
    std::unique_ptr<zetasql::ParserOutput> output;
    zetasql::ErrorMessageOptions err_opts;
    err_opts.mode = zetasql::ERROR_MESSAGE_WITH_PAYLOAD;
    auto s = zetasql::ParseScript(sql, opts, err_opts, &output);
    set_status(status, s);
}

void zetasql_AnalyzeStatement(
    const char* sql, void* catalog, void* opts, zetasql_Status* status) {
    std::unique_ptr<const zetasql::AnalyzerOutput> output;
    auto s = zetasql::AnalyzeStatement(
        sql,
        *static_cast<zetasql::AnalyzerOptions*>(opts),
        static_cast<zetasql::SimpleCatalog*>(catalog),
        static_cast<zetasql::SimpleCatalog*>(catalog)->type_factory(),
        &output);
    set_status(status, s);
}

void zetasql_free_string(char* s) {
    free(s);
}

} // extern "C"
