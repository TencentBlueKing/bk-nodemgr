.PHONY: all build clean test lint platform-builds build-matrix help

# 版本信息
BUILDTIME ?= $(shell date +%Y-%m-%dT%T%z)
GITTAG    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
GITHASH   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION   ?= ${GITTAG}-$(shell date +%y.%m.%d)

# 基本项目信息
APP_NAME    := installer
CURRENT_DIR := $(CURDIR)
MAIN_PATH   := $(CURDIR)/cmd/installer
OUTPUT_DIR  := $(CURDIR)/build/$(VERSION)

# Go构建标志
GO_FLAGS := -trimpath
LD_FLAGS := -ldflags "-s -w -X 'main.Version=$(VERSION)' -X 'main.CommitSHA=$(GITHASH)' -X 'main.BuildTime=$(BUILDTIME)'"

# 支持的操作系统
OSES := linux darwin windows

# 支持的架构
ARCHES := amd64 arm64 arm

# UPX压缩配置
UPX_ENABLED ?=
UPX_LEVEL   ?= 9
UPX_ARGS    ?= -$(UPX_LEVEL)

# fixed go version.
GO = CGO_ENABLED=0 go1.20.14

# cmd
MKDIR = mkdir -p
ECHO  = $(if $(filter Linux,$(shell uname)),echo -e,echo)
CD    = cd
CP    = cp
RM    = rm
SH    = sh

default: all

pre:
	go get golang.org/dl/go1.20.14@latest
	go install golang.org/dl/go1.20.14@latest
	go1.20.14 download

	@$(MAKE) check-upx

	@$(MKDIR) $(OUTPUT_DIR)

all: test build

# 构建本地平台的应用
build: | pre
	@$(ECHO) "Building $(APP_NAME) $(VERSION) for host platform..."
	$(GO) build $(GO_FLAGS) $(LD_FLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) $(MAIN_PATH)
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/$(APP_NAME)"

	@if [ "$(UPX_ENABLED)" ]; then \
		$(MAKE) compress-binary BINARY=$(OUTPUT_DIR)/$(APP_NAME); \
	fi

# 跨平台构建所有组合
platform-builds: | pre
	@$(ECHO) "Building $(APP_NAME) $(VERSION) for all platforms..."
	@$(foreach os,$(OSES),\
		$(foreach arch,$(ARCHES),\
			if [ "$(os)" = "darwin" ] && [ "$(arch)" = "arm" ]; then \
				$(ECHO) "Skipping unsupported platform: $(os)/$(arch)"; \
			else \
				$(ECHO) "Building for $(os)/$(arch)..." && \
				ext=$(if $(filter windows,$(os)),.exe,) && \
				binary="$(OUTPUT_DIR)/$(APP_NAME)_$(os)_$(arch)$$ext" && \
				GOOS=$(os) GOARCH=$(arch) $(GO) build $(GO_FLAGS) $(LD_FLAGS) \
					-o "$$binary" $(MAIN_PATH) && \
				$(ECHO) "Built: $$binary" && \
				if [ "$(UPX_ENABLED)" ]; then \
					$(MAKE) compress-binary BINARY=$$binary; \
				fi || exit 1; \
			fi; \
		)\
	)

# 压缩二进制文件
build-matrix: | pre
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		$(ECHO) "Error: GOOS and GOARCH must be specified."; \
		$(ECHO) "Example: make build-matrix GOOS=linux GOARCH=amd64"; \
		exit 1; \
	fi

	@$(ECHO) "Building $(APP_NAME) $(VERSION) for $(GOOS)/$(GOARCH)..."
	@ext=$(if $(filter windows,$(GOOS)),.exe,) && \
	binary="$(OUTPUT_DIR)/$(APP_NAME)_$(GOOS)_$(GOARCH)$$ext" && \
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GO_FLAGS) $(LD_FLAGS) \
		-o "$$binary" $(MAIN_PATH)
	@$(ECHO) "Built: $$binary"

	@if [ "$(UPX_ENABLED)" ]; then \
		$(MAKE) compress-binary BINARY="$(OUTPUT_DIR)/$(APP_NAME)_$(GOOS)_$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,)"; \
	fi

# 运行测试
test: | pre
	@$(ECHO) "Running tests..."

	@GOOS=$$($(GO) env GOOS); \
	GOARCH=$$($(GO) env GOARCH); \
	if [ "$$GOOS" = "windows" ] && [ "$$GOARCH" = "386" ]; then \
		$(ECHO) "Race detector disabled on windows/386"; \
		$(GO) test -coverprofile=$(OUTPUT_DIR)/coverage.out ./...; \
	else \
		$(GO) test -race -coverprofile=$(OUTPUT_DIR)/coverage.out ./...; \
	fi
	@if [ -f "$(OUTPUT_DIR)/coverage.out" ]; then \
		$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out; \
	fi
	@$(ECHO) "Tests passed"

# 清理构建文件
clean:
	@$(ECHO) "Cleaning tools build directory..."
	@$(RM) -rf build
	@$(ECHO) "Cleaned tools build directory"

# 代码格式检查
lint:
	@$(ECHO) "Checking code format..."
	$(GO)fmt -d -s .
	@$(ECHO) "Code format check completed"

# 帮助信息
help:
	@$(ECHO) "Go跨平台构建Makefile帮助 (GitHub Actions优化):"
	@$(ECHO) "基本命令:"
	@$(ECHO) "  make               - 清理、测试并构建当前平台的应用"
	@$(ECHO) "  make build         - 构建当前平台的应用"
	@$(ECHO) "  make platform-builds - 构建所有支持平台的应用"
	@$(ECHO) "  make clean         - 清理构建目录"
	@$(ECHO) "  make test          - 运行测试"
	@$(ECHO) "  make lint          - 检查代码格式"
	@$(ECHO) ""
	@$(ECHO) "GitHub Actions专用命令:"
	@$(ECHO) "  make build-matrix GOOS=linux GOARCH=amd64 - 构建指定平台的应用 (适用于矩阵构建)"
	@$(ECHO) "  make prepare-artifacts - 准备GitHub Actions制品目录"
	@$(ECHO) ""
	@$(ECHO) "环境变量:"
	@$(ECHO) "  VERSION           - 设置版本号 (默认: git tag或'v0.0.0')"
	@$(ECHO) ""
	@$(ECHO) "支持的平台组合:"
	@$(foreach os,$(OSES),\
		$(foreach arch,$(ARCHES),\
			if [ "$(os)" = "darwin" ] && [ "$(arch)" = "arm" ]; then \
				$(ECHO) "  ✗ $(os)/$(arch) (不支持)"; \
			else \
				$(ECHO) "  ✓ $(os)/$(arch)"; \
			fi; \
		)\
	)

# 检查UPX是否可用
check-upx:
	@if [ "$(UPX_ENABLED)" ]; then \
		if command -v upx >/dev/null 2>&1; then \
			$(ECHO) "UPX compression enabled (level $(UPX_LEVEL))"; \
		else \
			$(ECHO) "Warning: UPX compression requested but upx not found in PATH"; \
			$(ECHO) "Install UPX or set UPX_ENABLED="; \
			exit 1; \
		fi; \
	fi

# 压缩单个二进制文件
compress-binary:
	@if [ -z "$(BINARY)" ]; then \
		$(ECHO) "Error: BINARY path not specified"; \
		$(ECHO) "Usage: make compress-binary BINARY=path/to/binary"; \
		exit 1; \
	fi
	@if [ ! -f "$(BINARY)" ]; then \
		$(ECHO) "Error: Binary file not found: $(BINARY)"; \
		exit 1; \
	fi
	@if [ -z "$(UPX_ENABLED)" ]; then \
		$(ECHO) "Compressing $(BINARY) with UPX..."; \
		upx $(UPX_ARGS) "$(BINARY)"; \
		$(ECHO) "Compression complete"; \
	else \
		$(ECHO) "UPX compression disabled. Set UPX_ENABLED=1 to enable"; \
	fi

# 压缩所有构建的二进制文件
compress-all: check-upx
	@if [ "$(UPX_ENABLED)" ]; then \
		$(ECHO) "Compressing all binaries in $(OUTPUT_DIR)..."; \
		find "$(OUTPUT_DIR)" -type f -name "$(APP_NAME)*" -not -path "*/\.*" | while read binary; do \
			if file "$$binary" | grep -q "executable" || file "$$binary" | grep -q "PE32"; then \
				$(ECHO) "Compressing $$binary"; \
				upx $(UPX_ARGS) "$$binary"; \
			fi; \
		done; \
		$(ECHO) "All binaries compressed"; \
	else \
		$(ECHO) "UPX compression disabled. Set UPX_ENABLED=1 to enable"; \
	fi