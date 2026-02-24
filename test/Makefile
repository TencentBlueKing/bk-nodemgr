.PHONY: all clean test build pre mock-server

# test group execution order: precheck runs first, others follow
TEST_GROUPS = precheck backend application file

# version
BUILDTIME := $(shell date +%Y-%m-%dT%T%z)
GITTAG    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
GITHASH   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION   ?= ${GITTAG}-$(shell date +%y.%m.%d)

# directories
TEST_OUTPUT = build
DATA_DIR    = $(CURDIR)/data

# go version
GO = go1.23.10

# ldflags
LDVersionFLAG = "-X github.com/TencentBlueKing/bk-nodemgr/pkg/version.VERSION=${VERSION} \
	-X github.com/TencentBlueKing/bk-nodemgr/pkg/version.BUILDTIME=${BUILDTIME} \
	-X github.com/TencentBlueKing/bk-nodemgr/pkg/version.GITHASH=${GITHASH} "

# cmd
MKDIR = mkdir -p
ECHO  = $(if $(filter Linux,$(shell uname)),echo -e,echo)
CD    = cd
CP    = cp
RM    = rm

default: all

pre:
	@$(ECHO) "Building test cases..."

build:
	@$(RM) -rf $(TEST_OUTPUT)
	@$(MKDIR) $(TEST_OUTPUT)
	@for group in $(TEST_GROUPS); do \
		for testfile in $$(find ./cases/$$group -name "*_test.go" 2>/dev/null | sort); do \
			dir=$$(dirname "$$testfile"); \
			rel=$${dir#./cases/}; \
			test_name=$$(basename "$$testfile" _test.go); \
			out="$(TEST_OUTPUT)/$$rel/$$test_name.test"; \
			$(MKDIR) $$(dirname "$$out"); \
			CGO_ENABLED=0 $(GO) test -v -c -o "$$out" "./$$dir" || exit 1; \
		done; \
	done

mock-server: pre
	@$(ECHO) "Building mock-server $(VERSION)..."
	@$(MKDIR) $(TEST_OUTPUT)
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(TEST_OUTPUT)/mock-server ./mock-server
	@$(ECHO) "Built successfully: $(TEST_OUTPUT)/mock-server"

all: pre build mock-server
	@$(ECHO) "Built successfully test cases and mock-server"

test:
	@for group in $(TEST_GROUPS); do \
		for testfile in $$(find $(TEST_OUTPUT)/$$group -type f -name "*.test" 2>/dev/null | sort); do \
			echo "RUN [$$group]: $$testfile"; \
			./$$testfile -data-dir=$(DATA_DIR) || exit 1; \
		done; \
	done

clean:
	@$(RM) -rf $(TEST_OUTPUT)
