.PHONY: all clean test build pre mock-server

# version
BUILDTIME := $(shell date +%Y-%m-%dT%T%z)
GITTAG    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
GITHASH   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION   ?= ${GITTAG}-$(shell date +%y.%m.%d)

# directories
TEST_OUTPUT = build

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
	@for dir in $$(find ./cases -type d -name '*' | grep -v '^\./cases$$'); do \
		[ -f "$$dir"/*_test.go ] 2>/dev/null || continue; \
		rel=$${dir#./cases/}; \
		service=$${rel%%/*}; \
		rest=$${rel#*/}; \
		out="$(TEST_OUTPUT)/$${service}/$$rest.test"; \
		$(MKDIR) $$(dirname "$$out"); \
		CGO_ENABLED=0 $(GO) test -v -c -o "$$out" "./$$dir" || exit 1; \
	done

mock-server: pre
	@$(ECHO) "Building mock-server $(VERSION)..."
	@$(MKDIR) $(TEST_OUTPUT)
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(TEST_OUTPUT)/mock-server ./mock-server
	@$(ECHO) "Built successfully: $(TEST_OUTPUT)/mock-server"

all: pre build mock-server
	@$(ECHO) "Built successfully test cases and mock-server"

test:
	@for testfile in $$(find $(TEST_OUTPUT) -type f -name "*.test" 2>/dev/null); do \
		echo "RUN TESTCASE: $$testfile"; \
		./$$testfile || exit 1; \
	done

clean:
	@$(RM) -rf $(TEST_OUTPUT)

