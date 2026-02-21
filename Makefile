GOOGLESQL_DIR := third_party/googlesql
LIB_DIR := lib
INCLUDE_DIR := $(LIB_DIR)/include

.PHONY: lib clean test build

# Build the static library from googlesql via Bazel
lib: $(LIB_DIR)/libzetasql.a

$(LIB_DIR)/libzetasql.a: $(GOOGLESQL_DIR)/MODULE.bazel
	@echo "==> Building googlesql static library via Bazel..."
	@mkdir -p $(LIB_DIR) $(INCLUDE_DIR)
	cd $(GOOGLESQL_DIR) && mkdir -p bigq && cp ../../build/BUILD.bigq bigq/BUILD
	cd $(GOOGLESQL_DIR) && bazel build --compilation_mode=opt --check_visibility=false --experimental_cc_static_library //bigq:zetasql_complete
	@echo "==> Copying static library..."
	chmod u+w $(LIB_DIR)/libzetasql.a 2>/dev/null || true
	cp $(GOOGLESQL_DIR)/bazel-bin/bigq/libzetasql_complete.a $(LIB_DIR)/libzetasql.a
	@echo "==> Copying ICU libraries..."
	cp $$(find -L $(GOOGLESQL_DIR)/bazel-bin/external -name 'libicuuc.a' -path '*/copy_icu/*' | head -1) $(LIB_DIR)/
	cp $$(find -L $(GOOGLESQL_DIR)/bazel-bin/external -name 'libicui18n.a' -path '*/copy_icu/*' | head -1) $(LIB_DIR)/
	cp $$(find -L $(GOOGLESQL_DIR)/bazel-bin/external -name 'libicudata.a' -path '*/copy_icu/*' | head -1) $(LIB_DIR)/
	@echo "==> Copying headers..."
	./scripts/collect_headers.sh $(GOOGLESQL_DIR) $(INCLUDE_DIR)
	@echo "==> Done. Static library at $(LIB_DIR)/libzetasql.a"

# Build Go binaries
build: lib
	CGO_ENABLED=1 go build ./...

# Run tests
test: lib
	CGO_ENABLED=1 go test ./...

# Build the CLI binary
go-bigq: lib
	CGO_ENABLED=1 go build -o bin/go-bigq ./cmd/bigq/

# Clean build artifacts
clean:
	rm -rf $(LIB_DIR)/*.a $(INCLUDE_DIR)
	cd $(GOOGLESQL_DIR) && rm -rf bigq && bazel clean 2>/dev/null || true
