.PHONY: tidy build test pre backend application file relay front docker-build-server all clean doc tools bintools scripts apigw-docs

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
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-backend $(ROOT_DIR)/cmd/backend/main.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-backend"

application: | pre
	@$(ECHO) "Building application $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-application $(ROOT_DIR)/cmd/application/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-application"

file: | pre
	@$(ECHO) "Building file $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-file $(ROOT_DIR)/cmd/file/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-file"

relay: | pre
	@$(ECHO) "Building proxy $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-relay $(ROOT_DIR)/cmd/relay/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-relay"

front: | pre
	@$(ECHO) "Building frontend..."
	@$(CD) $(ROOT_DIR)/front && $(NPM) i && $(NPM) build
	@$(CP) -R $(ROOT_DIR)/front/dist $(OUTPUT_DIR)/
	@$(ECHO) "Built successfully: frontend"

tools: | pre
	@$(ECHO) "Building tools..."
	@$(MAKE) -j -C $(ROOT_DIR)/tools platform-builds -e UPX_ENABLED=1

	$(MKDIR) $(OUTPUT_DIR)/tools
	@$(CP) -r $(ROOT_DIR)/tools/build/$(VERSION)/* $(OUTPUT_DIR)/tools
	@$(ECHO) "Built successfully tools"

bintools: | pre
	@$(ECHO) "Building bintools..."
	@$(MKDIR) $(OUTPUT_DIR)/bintools/

	@$(ECHO) "Building gsectl bintool..."
	@$(MAKE) -C $(ROOT_DIR)/script_tools/gsectl
	@$(CP) -r $(ROOT_DIR)/script_tools/gsectl/build/$(VERSION)/bintool.tgz $(OUTPUT_DIR)/bintools/

	@$(ECHO) "Building plugin bintool..."
	@$(MAKE) -C $(ROOT_DIR)/script_tools/plugin_scripts
	@$(CP) -r $(ROOT_DIR)/script_tools/plugin_scripts/build/$(VERSION)/plugin_bintool.tgz $(OUTPUT_DIR)/bintools/

	@$(ECHO) "Built successfully $(OUTPUT_DIR)/bintools/bintool.tgz"
	@$(ECHO) "Built successfully $(OUTPUT_DIR)/bintools/plugin_bintool.tgz"
	@$(ECHO) "Built successfully bintools"

scripts: | pre
	@$(ECHO) "Building scripts..."
	@$(MKDIR) $(OUTPUT_DIR)/scripts

	@$(CP) -R $(ROOT_DIR)/script_tools/manual $(OUTPUT_DIR)/scripts/

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
				GOOS=$(os) GOARCH=$(arch) $(GO) build $(GO_FLAGS) -ldflags $(LDVersionFLAG) \
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

docker-build-server: backend application file front tools scripts
	@$(ECHO) "Building docker images..."
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr/Dockerfile $(OUTPUT_DIR)
	@$(CP) $(ROOT_DIR)/install/docker-compose/serviced.sh $(OUTPUT_DIR)
	@$(CD) $(OUTPUT_DIR) && docker build -t bk-nodemgr-server:v${VERSION} .
	@$(ECHO) "Built successfully docker images bk-nodemgr-server:v${VERSION}"

docker-build-apigw-sync: | pre
	@$(ECHO) "Building docker image bk-nodemgr-apigw-sync..."
	@$(MKDIR) $(OUTPUT_DIR)/apigw-sync
	@$(CP) -R $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/support-files $(OUTPUT_DIR)/apigw-sync/
	@$(CP) -R $(ROOT_DIR)/apigw/* $(OUTPUT_DIR)/apigw-sync/support-files
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/Dockerfile $(OUTPUT_DIR)/apigw-sync
	@$(CD) $(OUTPUT_DIR)/apigw-sync && docker build -t bk-nodemgr-apigw-sync:v${VERSION} .
	@$(ECHO) "Built successfully docker image bk-nodemgr-apigw-sync:v${VERSION}"

test: | pre
	@$(ECHO) "Building test..."
	@$(MAKE) -C $(ROOT_DIR)/test build
	@$(MKDIR) $(OUTPUT_DIR)/test
	@$(CP) -R $(ROOT_DIR)/test/build/* $(OUTPUT_DIR)/test/ 2>/dev/null || true
	@$(ECHO) "Built successfully test"

all: backend application file relay front tools scripts bintools test

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
	@$(CD) $(ROOT_DIR) && golangci-lint run --config ${ROOT_DIR}/.golangci.yml --path-prefix ${ROOT_DIR}
	@$(ECHO) "Linting completed"