#!/usr/bin/env bash
set -euo pipefail

# Copy headers needed for compiling the C bridge from googlesql source and
# Bazel-generated files (protos, etc.) into lib/include/.
# Usage: collect_headers.sh <googlesql_dir> <include_dir>

GOOGLESQL_DIR="$1"
INCLUDE_DIR="$(cd "$(dirname "$2")" && pwd)/$(basename "$2")"

mkdir -p "$INCLUDE_DIR"

copy_with_parents() {
    local src="$1"
    local base="$2"
    local dest="$3"
    local rel="${src#${base}/}"
    mkdir -p "$dest/$(dirname "$rel")"
    cp -f "$src" "$dest/$rel"
}

# 1. Copy googlesql source headers
echo "Copying googlesql source headers..."
cd "$GOOGLESQL_DIR"
find googlesql -name '*.h' -type f | while read -r f; do
    copy_with_parents "$f" "." "$INCLUDE_DIR"
done

# 2. Copy Bazel-generated headers (.pb.h, flex/bison output, etc.)
BAZEL_BIN="bazel-bin"
if [ -d "$BAZEL_BIN" ]; then
    echo "Copying Bazel-generated headers..."
    find -L "$BAZEL_BIN" -name '*.h' -path '*/googlesql/*' -type f 2>/dev/null | while read -r f; do
        rel="${f#${BAZEL_BIN}/}"
        mkdir -p "$INCLUDE_DIR/$(dirname "$rel")"
        cp -f "$f" "$INCLUDE_DIR/$rel"
    done
fi

# 3. Copy external dependency headers
OUTPUT_BASE=$(bazel info output_base 2>/dev/null || true)
if [ -z "$OUTPUT_BASE" ]; then
    echo "WARNING: Could not determine Bazel output base."
    cd - >/dev/null
    exit 0
fi

copy_external_headers() {
    local name="$1"
    local subdir="$2"  # subdirectory within the external repo to copy from (e.g. "src" for protobuf)
    local pattern="$3" # find pattern (e.g. "*.h" or "*.inc")

    local repo_path=""
    # Try exact match first, then prefix match
    for candidate in "$OUTPUT_BASE/external/$name" "$OUTPUT_BASE/external/${name}~"; do
        if [ -d "$candidate" ]; then
            repo_path="$candidate"
            break
        fi
    done

    if [ -z "$repo_path" ]; then
        # Try fuzzy match
        repo_path=$(find "$OUTPUT_BASE/external" -maxdepth 1 -name "${name}*" -type d 2>/dev/null | head -1)
    fi

    if [ -n "$repo_path" ] && [ -d "$repo_path" ]; then
        local search_dir="$repo_path"
        if [ -n "$subdir" ] && [ -d "$repo_path/$subdir" ]; then
            search_dir="$repo_path/$subdir"
        fi

        find "$search_dir" -name "$pattern" -type f 2>/dev/null | while read -r f; do
            local rel="${f#${search_dir}/}"
            mkdir -p "$INCLUDE_DIR/$(dirname "$rel")"
            cp -f "$f" "$INCLUDE_DIR/$rel"
        done
        echo "  Copied $name from $search_dir"
    else
        echo "  WARNING: $name not found in $OUTPUT_BASE/external/"
    fi
}

echo "Copying external dependency headers..."

# Abseil
copy_external_headers "abseil-cpp" "" "*.h"
copy_external_headers "abseil-cpp" "" "*.inc"

# Protobuf
copy_external_headers "protobuf" "src" "*.h"
copy_external_headers "protobuf" "src" "*.inc"

# Google Test (just gtest_prod.h for FRIEND_TEST)
copy_external_headers "googletest" "googletest/include" "*.h"

cd - >/dev/null

# Fix permissions (Bazel outputs may be read-only)
chmod -R u+w "$INCLUDE_DIR" 2>/dev/null || true

HEADER_COUNT=$(find "$INCLUDE_DIR" -name '*.h' -o -name '*.inc' -type f | wc -l | tr -d ' ')
echo "Copied $HEADER_COUNT header/include files to $INCLUDE_DIR"
