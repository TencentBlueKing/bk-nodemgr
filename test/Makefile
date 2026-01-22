.PHONY: all clean test build pre

# directories
TEST_OUTPUT = build

# go version
GO = go1.23.10

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

all: pre build
	@$(ECHO) "Built successfully test cases"

test:
	@for testfile in $$(find $(TEST_OUTPUT) -type f -name "*.test" 2>/dev/null); do \
		echo "RUN TESTCASE: $$testfile"; \
		./$$testfile || exit 1; \
	done

clean:
	@$(RM) -rf $(TEST_OUTPUT)

