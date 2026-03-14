.PHONY: tidy build test pre backend application file relay front mock-server docker-build-server docker-build-mock-server all clean doc tools bintools scripts apigw-docs support-files

# Target platform for docker-build-server (optional)
# Examples:
#   make docker-build-server                      # build for host arch
#   make docker-build-server TARGET_PLATFORM=linux/amd64
#   make docker-build-server TARGET_PLATFORM=linux/arm64
# Notes:
#   - When TARGET_PLATFORM is set, service binaries are cross-compiled for that platform
#     and the docker image is built via buildx with the same --platform.
TARGET_PLATFORM ?=

# buildx builder (used when TARGET_PLATFORM is set)
BUILDX_BUILDER ?= nodemgr-multiarch
BUILDX_PLATFORMS ?= linux/amd64,linux/arm64

ifneq ($(strip $(TARGET_PLATFORM)),)
TARGET_OS := $(word 1,$(subst /, ,$(TARGET_PLATFORM)))
TARGET_ARCH := $(word 2,$(subst /, ,$(TARGET_PLATFORM)))
endif

BASE_IMAGE ?= alpine

# version
BUILDTIME := $(shell date +%Y-%m-%dT%T%z)
GITTAG    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
GITHASH   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION   ?= ${GITTAG}-$(shell date +%y.%m.%d)

# directories
ROOT_DIR   := $(CURDIR)
OUTPUT_DIR := $(ROOT_DIR)/build/$(VERSION)

# ldflags
# output directory for release package and version for command line
LDVersionFLAG = "-X github.com/TencentBlueKing/bk-nodemgr/pkg/version.VERSION=${VERSION} \
	-X github.com/TencentBlueKing/bk-nodemgr/pkg/version.BUILDTIME=${BUILDTIME} \
	-X github.com/TencentBlueKing/bk-nodemgr/pkg/version.GITHASH=${GITHASH} \
	-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore"

# fixed go version.
GO = go1.23.10

# Go build env for cross compile (empty when TARGET_PLATFORM is not set)
GO_BUILD_ENV = CGO_ENABLED=0 $(if $(TARGET_OS),GOOS=$(TARGET_OS),) $(if $(TARGET_ARCH),GOARCH=$(TARGET_ARCH),)

# cmd
MKDIR = mkdir -p
ECHO  = $(if $(filter Linux,$(shell uname)),echo -e,echo)
SED   = sed
CD    = cd
CP    = cp
RM    = rm
SH    = sh
TAR   = tar -zcf
NPM   = pnpm

default: all

pre:
	@$(MKDIR) $(OUTPUT_DIR)

	go get golang.org/dl/go1.23.10@latest
	go install golang.org/dl/go1.23.10@latest
	go1.23.10 download

	go1.23.10 mod tidy

backend: | pre
	@$(ECHO) "Building backend $(VERSION)..."
	$(GO_BUILD_ENV) $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-backend $(ROOT_DIR)/cmd/backend/main.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-backend"

application: | pre
	@$(ECHO) "Building application $(VERSION)..."
	$(GO_BUILD_ENV) $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-application $(ROOT_DIR)/cmd/application/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-application"

file: | pre
	@$(ECHO) "Building file $(VERSION)..."
	$(GO_BUILD_ENV) $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-file $(ROOT_DIR)/cmd/file/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-file"

relay: | pre
	@$(ECHO) "Building proxy $(VERSION)..."
	$(GO_BUILD_ENV) $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-relay $(ROOT_DIR)/cmd/relay/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-relay"

mock-server: | pre
	@$(ECHO) "Building mock-server $(VERSION)..."
	$(GO_BUILD_ENV) $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/mock-server $(ROOT_DIR)/test/mock-server/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/mock-server"

front: | pre
	@$(ECHO) "Building frontend..."
	@$(CD) $(ROOT_DIR)/front && $(NPM) i && $(NPM) build
	@$(CP) -R $(ROOT_DIR)/front/dist $(OUTPUT_DIR)/
	@$(ECHO) "Built successfully: frontend"

tools: | pre
	@$(ECHO) "Building tools..."
	@$(MAKE) -j -C $(ROOT_DIR)/tools platform-builds VERSION="$(VERSION)" -e UPX_ENABLED=1

	$(MKDIR) $(OUTPUT_DIR)/tools
	@$(CP) -r $(ROOT_DIR)/tools/build/$(VERSION)/* $(OUTPUT_DIR)/tools
	@$(ECHO) "Built successfully tools"

bintools: | pre
	@$(ECHO) "Building bintools..."
	@$(MKDIR) $(OUTPUT_DIR)/bintools/

	@$(ECHO) "Building gsectl bintool..."
	@$(MAKE) -C $(ROOT_DIR)/script_tools/gsectl VERSION="$(VERSION)"
	@$(CP) -r $(ROOT_DIR)/script_tools/gsectl/build/$(VERSION)/bintool.tgz $(OUTPUT_DIR)/bintools/

	@$(ECHO) "Building plugin bintool..."
	@$(MAKE) -C $(ROOT_DIR)/script_tools/plugin_scripts VERSION="$(VERSION)"
	@$(CP) -r $(ROOT_DIR)/script_tools/plugin_scripts/build/$(VERSION)/plugin_bintool.tgz $(OUTPUT_DIR)/bintools/

	@$(ECHO) "Built successfully $(OUTPUT_DIR)/bintools/bintool.tgz"
	@$(ECHO) "Built successfully $(OUTPUT_DIR)/bintools/plugin_bintool.tgz"
	@$(ECHO) "Built successfully bintools"

scripts: | pre
	@$(ECHO) "Building scripts..."
	@$(MKDIR) $(OUTPUT_DIR)/scripts

	@$(CP) -R $(ROOT_DIR)/script_tools/manual $(OUTPUT_DIR)/scripts/
	@$(CP) -R $(ROOT_DIR)/script_tools/ops $(OUTPUT_DIR)/scripts/

support-files: | pre
	@$(ECHO) "Building support-files..."
	@$(CP) -R $(ROOT_DIR)/support-files $(OUTPUT_DIR)
	$(GO_BUILD_ENV) $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/support-files/initpackage/jwt-generator $(ROOT_DIR)/support-files/initpackage/jwt_generator/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/support-files/initpackage/jwt-generator"

OSES := linux

# 支持的架构
ARCHES := amd64 arm64

plugin-pkg-relay: APP_NAME ?= bk-nodemgr-relay
plugin-pkg-relay: | pre
	@$(ECHO) "Building plugin-pkg-relay..."
	@$(MKDIR) $(OUTPUT_DIR)/$(APP_NAME)
	@$(ECHO) "Building $(APP_NAME) $(VERSION) for all platforms..."
	@$(foreach os,$(OSES),\
		$(foreach arch,$(ARCHES),\
			if [ "$(os)" = "darwin" ] && [ "$(arch)" = "arm" ]; then \
				$(ECHO) "Skipping unsupported platform: $(os)/$(arch)"; \
			else \
				$(ECHO) "Building for $(os)/$(arch)..." && \
				ext=$(if $(filter windows,$(os)),.exe,) && \
				plugin_path=$(OUTPUT_DIR)/$(APP_NAME)/plugins_$(os)_$(arch); \
                if [ "$(os)" = "linux" ] && [ "$(arch)" = "amd64" ]; then \
                    plugin_path=$(OUTPUT_DIR)/$(APP_NAME)/plugins_linux_x86_64; \
                else \
                    plugin_path=$(OUTPUT_DIR)/$(APP_NAME)/plugins_linux_aarch64; \
                fi; \
				$(MKDIR) $$plugin_path && \
				$(MKDIR) $$plugin_path/bin && \
				binary="$$plugin_path/bin/$(APP_NAME)$$ext" && \
				CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) $(GO) build $(GO_FLAGS) -ldflags $(LDVersionFLAG) \
					-o $$binary $(ROOT_DIR)/cmd/relay/*.go && \
				$(ECHO) "Built: $$binary" && \
				if [ "$(UPX_ENABLED)" ]; then \
					$(MAKE) compress-binary BINARY=$$binary; \
				fi; \
				$(MKDIR) "$$plugin_path/etc"; \
				$(MKDIR) "$$plugin_path/templates"; \
				$(CP) $(ROOT_DIR)/plugin/relay/definition.yaml "$$plugin_path/definition.yaml"; \
				$(CP) $(ROOT_DIR)/plugin/relay/templates/*.template "$$plugin_path/templates"; \
			fi; \
		)\
	)

	@$(ECHO) "Rendering project.yaml with VERSION=$(VERSION)..."
	@$(SED) 's/{{VERSION}}/$(VERSION)/g' "./plugin/relay/project.yaml" > "$(OUTPUT_DIR)/$(APP_NAME)/project.yaml"

	@$(ECHO) "Packaging artifacts..."
	@$(TAR) "$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION).tgz" -C "$(OUTPUT_DIR)" $(APP_NAME)

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

docker-build-server: backend application file front tools scripts bintools support-files
	@$(ECHO) "Building docker images..."
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr/${BASE_IMAGE}/Dockerfile $(OUTPUT_DIR)
	@$(CP) $(ROOT_DIR)/install/docker-compose/bk-nodemgr/serviced.sh $(OUTPUT_DIR)
	@if [ -n "$(TARGET_PLATFORM)" ]; then \
		case "$(TARGET_PLATFORM)" in \
			*/*) ;; \
			*) $(ECHO) "Error: TARGET_PLATFORM must be in 'os/arch' format, got '$(TARGET_PLATFORM)'"; exit 1;; \
		esac; \
		$(ECHO) "Make sure Docker buildx is available..."; \
		docker buildx version >/dev/null 2>&1 || { $(ECHO) "Please install/enable Docker buildx"; exit 1; }; \
		prev_builder=$$(docker buildx ls 2>/dev/null | awk '$$1 ~ /\*/ {gsub("\\*","",$$1); print $$1; exit}'); \
		[ -n "$$prev_builder" ] || prev_builder=default; \
		trap 'docker buildx use default >/dev/null 2>&1 || docker buildx use "$$prev_builder" >/dev/null 2>&1 || true' EXIT; \
		if ! docker buildx inspect "$(BUILDX_BUILDER)" >/dev/null 2>&1; then \
			$(ECHO) "Creating buildx builder: $(BUILDX_BUILDER)"; \
			docker buildx create --name "$(BUILDX_BUILDER)" --driver docker-container --platform "$(BUILDX_PLATFORMS)" --use --bootstrap; \
		else \
			$(ECHO) "Using existing buildx builder: $(BUILDX_BUILDER)"; \
			docker buildx use "$(BUILDX_BUILDER)"; \
			docker buildx inspect --bootstrap >/dev/null; \
		fi; \
		$(ECHO) "Building docker image for platform $(TARGET_PLATFORM)..."; \
		$(CD) $(OUTPUT_DIR) && docker buildx build --platform $(TARGET_PLATFORM) -t bk-nodemgr-server:v${VERSION} --load .; \
	else \
		$(CD) $(OUTPUT_DIR) && docker build -t bk-nodemgr-server:v${VERSION} .; \
	fi
	@$(ECHO) "Built successfully docker images bk-nodemgr-server:v${VERSION}"

docker-build-apigw-sync: | pre
	@$(ECHO) "Building docker image bk-nodemgr-apigw-sync..."
	@$(MKDIR) $(OUTPUT_DIR)/apigw-sync
	@$(CP) -R $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/support-files $(OUTPUT_DIR)/apigw-sync/
	@$(CP) -R $(ROOT_DIR)/apigw/* $(OUTPUT_DIR)/apigw-sync/support-files
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/Dockerfile $(OUTPUT_DIR)/apigw-sync
	@$(CD) $(OUTPUT_DIR)/apigw-sync && docker build -t bk-nodemgr-apigw-sync:v${VERSION} .
	@$(ECHO) "Built successfully docker image bk-nodemgr-apigw-sync:v${VERSION}"

docker-build-mock-server: mock-server
	@$(ECHO) "Building docker image mock-server..."
	@$(MKDIR) $(OUTPUT_DIR)/mock-server-image
	@$(CP) $(OUTPUT_DIR)/mock-server $(OUTPUT_DIR)/mock-server-image/mock-server
	@$(CP) $(ROOT_DIR)/install/images/mock-server/Dockerfile $(OUTPUT_DIR)/mock-server-image/
	@if [ -n "$(TARGET_PLATFORM)" ]; then \
		case "$(TARGET_PLATFORM)" in \
			*/*) ;; \
			*) $(ECHO) "Error: TARGET_PLATFORM must be in 'os/arch' format, got '$(TARGET_PLATFORM)'"; exit 1;; \
		esac; \
		$(ECHO) "Make sure Docker buildx is available..."; \
		docker buildx version >/dev/null 2>&1 || { $(ECHO) "Please install/enable Docker buildx"; exit 1; }; \
		prev_builder=$$(docker buildx ls 2>/dev/null | awk '$$1 ~ /\*/ {gsub("\\*","",$$1); print $$1; exit}'); \
		[ -n "$$prev_builder" ] || prev_builder=default; \
		trap 'docker buildx use default >/dev/null 2>&1 || docker buildx use "$$prev_builder" >/dev/null 2>&1 || true' EXIT; \
		if ! docker buildx inspect "$(BUILDX_BUILDER)" >/dev/null 2>&1; then \
			$(ECHO) "Creating buildx builder: $(BUILDX_BUILDER)"; \
			docker buildx create --name "$(BUILDX_BUILDER)" --driver docker-container --platform "$(BUILDX_PLATFORMS)" --use --bootstrap; \
		else \
			$(ECHO) "Using existing buildx builder: $(BUILDX_BUILDER)"; \
			docker buildx use "$(BUILDX_BUILDER)"; \
			docker buildx inspect --bootstrap >/dev/null; \
		fi; \
		$(ECHO) "Building docker image for platform $(TARGET_PLATFORM)..."; \
		$(CD) $(OUTPUT_DIR)/mock-server-image && docker buildx build --platform $(TARGET_PLATFORM) -t mock-server:v${VERSION} --load .; \
	else \
		$(CD) $(OUTPUT_DIR)/mock-server-image && docker build -t mock-server:v${VERSION} .; \
	fi
	@$(ECHO) "Built successfully docker image mock-server:v${VERSION}"
	@rm -rf $(OUTPUT_DIR)/mock-server-image

test: | pre
	@$(ECHO) "Building test..."
	@$(MAKE) -C $(ROOT_DIR)/test all
	@$(MKDIR) $(OUTPUT_DIR)/test
	@$(CP) -R $(ROOT_DIR)/test/build/* $(OUTPUT_DIR)/test/ 2>/dev/null || true
	@$(ECHO) "Built successfully test"

all: backend application file relay front tools scripts bintools support-files test

clean:
	@$(ECHO) "Cleaning build directory..."
	@$(RM) -rf build
	$(MAKE) -C $(ROOT_DIR)/tools clean
	$(MAKE) -C $(ROOT_DIR)/test clean
	@$(ECHO) "Cleaned build directory"

doc:
	@$(ECHO) "Open http://localhost:6060 to view the documentation"
	godoc -http=localhost:6060

apigw-docs: | pre
	@$(ECHO) "Building apigw-docs..."
	$(MKDIR) $(OUTPUT_DIR)/apigw
	@$(CD) $(ROOT_DIR)/apigw/apidocs && $(TAR) $(OUTPUT_DIR)/apigw/apidocs.tgz zh/ en/
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/apigw-docs/docs.tgz"

lint: | pre
	@$(ECHO) "Linting..."
	@$(CD) $(ROOT_DIR) && GOGC=40 $(if $(strip $(GOMEMLIMIT)),GOMEMLIMIT=$(GOMEMLIMIT),) golangci-lint run --config $(ROOT_DIR)/.golangci.yml --path-prefix $(ROOT_DIR)
	@$(ECHO) "Linting completed"
