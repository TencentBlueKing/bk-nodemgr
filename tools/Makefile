# 基本项目信息
APP_NAME := installer
MAIN_PATH := ./cmd/installer/
OUTPUT_DIR := ./bin

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
COMMIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Go构建标志
GO_FLAGS := -trimpath
LD_FLAGS := -ldflags "-s -w -X 'main.Version=$(VERSION)' -X 'main.CommitSHA=$(COMMIT_SHA)' -X 'main.BuildTime=$(BUILD_TIME)'"

# 支持的操作系统
OSES := linux darwin windows
# 支持的架构
ARCHES := amd64 arm64 arm

GO = go1.20.14

.PHONY: all build clean test lint platform-builds build-matrix help

# 默认目标
all: clean test build

# 构建本地平台的应用
build: check-upx
	@echo "Building $(APP_NAME) $(VERSION) for host platform..."
	@mkdir -p $(OUTPUT_DIR)
	@$(GO) build $(GO_FLAGS) $(LD_FLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build successful: $(OUTPUT_DIR)/$(APP_NAME)"
	@if [ "$(UPX_ENABLED)" = "true" ]; then \
		make compress-binary BINARY=$(OUTPUT_DIR)/$(APP_NAME); \
	fi

# 跨平台构建所有组合（支持UPX压缩）
platform-builds: check-upx
	@echo "Building $(APP_NAME) $(VERSION) for all platforms..."
	@mkdir -p $(OUTPUT_DIR)
	@$(foreach os,$(OSES),\
		$(foreach arch,$(ARCHES),\
			if [ "$(os)" = "darwin" ] && [ "$(arch)" = "arm" ]; then \
				echo "Skipping unsupported platform: $(os)/$(arch)"; \
			else \
				echo "Building for $(os)/$(arch)..." && \
				ext=$(if $(filter windows,$(os)),.exe,) && \
				binary="$(OUTPUT_DIR)/$(APP_NAME)_$(os)_$(arch)$$ext" && \
				GOOS=$(os) GOARCH=$(arch) $(GO) build $(GO_FLAGS) $(LD_FLAGS) \
					-o "$$binary" $(MAIN_PATH) && \
				echo "Built: $$binary" && \
				if [ "$(UPX_ENABLED)" = "true" ]; then \
					make compress-binary BINARY=$$binary; \
				fi || exit 1; \
			fi; \
		)\
	)

# 压缩二进制文件
build-matrix: check-upx
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "Error: GOOS and GOARCH must be specified."; \
		echo "Example: make build-matrix GOOS=linux GOARCH=amd64"; \
		exit 1; \
	fi
	@echo "Building $(APP_NAME) $(VERSION) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(OUTPUT_DIR)
	@ext=$(if $(filter windows,$(GOOS)),.exe,) && \
	binary="$(OUTPUT_DIR)/$(APP_NAME)_$(GOOS)_$(GOARCH)$$ext" && \
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GO_FLAGS) $(LD_FLAGS) \
		-o "$$binary" $(MAIN_PATH) && \
	echo "Built: $$binary"
	@if [ "$(UPX_ENABLED)" = "true" ]; then \
		make compress-binary BINARY="$(OUTPUT_DIR)/$(APP_NAME)_$(GOOS)_$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,)"; \
	fi

# 清理构建文件
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(OUTPUT_DIR)
	@mkdir -p $(OUTPUT_DIR)
	@echo "Cleaned build directory"

# 运行测试
test:
	@echo "Running tests..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=$$($(GO) env GOOS); \
	GOARCH=$$($(GO) env GOARCH); \
	if [ "$$GOOS" = "windows" ] && [ "$$GOARCH" = "386" ]; then \
		echo "Race detector disabled on windows/386"; \
		$(GO) test -coverprofile=$(OUTPUT_DIR)/coverage.out ./...; \
	else \
		$(GO) test -race -coverprofile=$(OUTPUT_DIR)/coverage.out ./...; \
	fi
	@if [ -f "$(OUTPUT_DIR)/coverage.out" ]; then \
		$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out; \
	fi
	@echo "Tests passed"

# 代码格式检查
lint:
	@echo "Checking code format..."
	@$(GO)fmt -d -s .
	@echo "Code format check completed"

# 帮助信息
help:
	@echo "Go跨平台构建Makefile帮助 (GitHub Actions优化):"
	@echo "基本命令:"
	@echo "  make               - 清理、测试并构建当前平台的应用"
	@echo "  make build         - 构建当前平台的应用"
	@echo "  make platform-builds - 构建所有支持平台的应用"
	@echo "  make clean         - 清理构建目录"
	@echo "  make test          - 运行测试"
	@echo "  make lint          - 检查代码格式"
	@echo ""
	@echo "GitHub Actions专用命令:"
	@echo "  make build-matrix GOOS=linux GOARCH=amd64 - 构建指定平台的应用 (适用于矩阵构建)"
	@echo "  make prepare-artifacts - 准备GitHub Actions制品目录"
	@echo ""
	@echo "环境变量:"
	@echo "  VERSION           - 设置版本号 (默认: git tag或'v0.0.0')"
	@echo ""
	@echo "支持的平台组合:"
	@$(foreach os,$(OSES),\
		$(foreach arch,$(ARCHES),\
			if [ "$(os)" = "darwin" ] && [ "$(arch)" = "arm" ]; then \
				echo "  ✗ $(os)/$(arch) (不支持)"; \
			else \
				echo "  ✓ $(os)/$(arch)"; \
			fi; \
		)\
	)

# UPX压缩配置
UPX_ENABLED ?= false
UPX_LEVEL ?= 9
UPX_ARGS ?= -$(UPX_LEVEL)

# 检查UPX是否可用
check-upx:
	@if [ "$(UPX_ENABLED)" = "true" ]; then \
		if command -v upx >/dev/null 2>&1; then \
			echo "UPX compression enabled (level $(UPX_LEVEL))"; \
		else \
			echo "Warning: UPX compression requested but upx not found in PATH"; \
			echo "Install UPX or set UPX_ENABLED=false"; \
			exit 1; \
		fi; \
	fi

# 压缩单个二进制文件
compress-binary:
	@if [ -z "$(BINARY)" ]; then \
		echo "Error: BINARY path not specified"; \
		echo "Usage: make compress-binary BINARY=path/to/binary"; \
		exit 1; \
	fi
	@if [ ! -f "$(BINARY)" ]; then \
		echo "Error: Binary file not found: $(BINARY)"; \
		exit 1; \
	fi
	@if [ "$(UPX_ENABLED)" = "true" ]; then \
		echo "Compressing $(BINARY) with UPX..."; \
		upx $(UPX_ARGS) "$(BINARY)"; \
		echo "Compression complete"; \
	else \
		echo "UPX compression disabled. Set UPX_ENABLED=true to enable"; \
	fi

# 压缩所有构建的二进制文件
compress-all: check-upx
	@if [ "$(UPX_ENABLED)" = "true" ]; then \
		echo "Compressing all binaries in $(OUTPUT_DIR)..."; \
		find "$(OUTPUT_DIR)" -type f -name "$(APP_NAME)*" -not -path "*/\.*" | while read binary; do \
			if file "$$binary" | grep -q "executable" || file "$$binary" | grep -q "PE32"; then \
				echo "Compressing $$binary"; \
				upx $(UPX_ARGS) "$$binary"; \
			fi; \
		done; \
		echo "All binaries compressed"; \
	else \
		echo "UPX compression disabled. Set UPX_ENABLED=true to enable"; \
	fi