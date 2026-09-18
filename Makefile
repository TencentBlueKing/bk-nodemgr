.PHONY: tidy build test pre backend application adminclient file relay front mock-server docker-build-server docker-build-mock-server all clean doc tools bintools scripts apigw-docs support-files helm

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
GO = go1.25.12
GOLANGCI_LINT_TIMEOUT ?= 2m

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
HELM  = helm

default: all

pre:
	@$(MKDIR) $(OUTPUT_DIR)

	go get golang.org/dl/go1.25.12@latest
	go install golang.org/dl/go1.25.12@latest
	go1.25.12 download

	go1.25.12 mod tidy

backend: | pre
	@$(ECHO) "Building backend $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-backend $(ROOT_DIR)/cmd/backend/main.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-backend"

application: | pre
	@$(ECHO) "Building application $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-application $(ROOT_DIR)/cmd/application/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-application"

adminclient: | pre
	@$(ECHO) "Building adminclient $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-adminclient $(ROOT_DIR)/cmd/adminclient/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-adminclient"

file: | pre
	@$(ECHO) "Building file $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-file $(ROOT_DIR)/cmd/file/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-file"

relay: | pre
	@$(ECHO) "Building proxy $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-relay $(ROOT_DIR)/cmd/relay/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-relay"

mock-server: | pre
	@$(ECHO) "Building mock-server $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/mock-server $(ROOT_DIR)/test/mock-server/*.go
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
	@$(MKDIR) $(OUTPUT_DIR)/support-files
	@for entry in $(ROOT_DIR)/support-files/* $(ROOT_DIR)/support-files/.[!.]* $(ROOT_DIR)/support-files/..?*; do \
		[ -e "$$entry" ] || [ -L "$$entry" ] || continue; \
		if [ "$${entry##*/}" != "bkiamv4" ]; then $(CP) -R "$$entry" $(OUTPUT_DIR)/support-files/ || exit 1; fi; \
	done
	@$(RM) -rf $(OUTPUT_DIR)/support-files/bkiamv4
	@$(MKDIR) $(OUTPUT_DIR)/support-files/bkiamv4/render $(OUTPUT_DIR)/support-files/bkiamv4/migrate $(OUTPUT_DIR)/support-files/bkiamv4/templates
	@$(CP) $(ROOT_DIR)/support-files/bkiamv4/templates/*.json.tpl $(OUTPUT_DIR)/support-files/bkiamv4/templates/
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/support-files/bkiamv4/render/iam-render ./support-files/bkiamv4/render
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/support-files/bkiamv4/migrate/iam-migrate ./support-files/bkiamv4/migrate
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/support-files/initpackage/jwt-generator $(ROOT_DIR)/support-files/initpackage/jwt_generator/*.go
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

docker-build-server: backend application file adminclient tools scripts bintools support-files
	@$(ECHO) "Building docker images..."
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr/${BASE_IMAGE}/Dockerfile $(OUTPUT_DIR)
	@$(CP) $(ROOT_DIR)/install/docker-compose/bk-nodemgr/serviced.sh $(OUTPUT_DIR)
	@$(ECHO) "Preparing frontend source for docker build..."
	@$(MKDIR) $(OUTPUT_DIR)/front
	@tar -C $(ROOT_DIR)/front --exclude=node_modules --exclude=dist -cf - . | tar -C $(OUTPUT_DIR)/front -xf -
	@$(CP) $(ROOT_DIR)/front/.dockerignore $(OUTPUT_DIR)/.dockerignore
	@$(ECHO) "Building docker image in docker build"
	@$(CD) $(OUTPUT_DIR) && docker build -t bk-nodemgr-server:${VERSION} .
	@$(ECHO) "Built successfully docker images bk-nodemgr-server:${VERSION}"

docker-build-apigw-sync: | pre
	@$(ECHO) "Building docker image bk-nodemgr-apigw-sync..."
	@$(MKDIR) $(OUTPUT_DIR)/apigw-sync
	@$(CP) -R $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/support-files $(OUTPUT_DIR)/apigw-sync/
	@$(CP) -R $(ROOT_DIR)/apigw/* $(OUTPUT_DIR)/apigw-sync/support-files
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/Dockerfile $(OUTPUT_DIR)/apigw-sync
	@$(CD) $(OUTPUT_DIR)/apigw-sync && docker build -t bk-nodemgr-apigw-sync:${VERSION} .
	@$(ECHO) "Built successfully docker image bk-nodemgr-apigw-sync:${VERSION}"

docker-build-mock-server: mock-server
	@$(ECHO) "Building docker image mock-server..."
	@$(MKDIR) $(OUTPUT_DIR)/mock-server-image
	@$(CP) $(OUTPUT_DIR)/mock-server $(OUTPUT_DIR)/mock-server-image/mock-server
	@$(CP) $(ROOT_DIR)/install/images/mock-server/Dockerfile $(OUTPUT_DIR)/mock-server-image/
	@$(CD) $(OUTPUT_DIR)/mock-server-image && docker build -t mock-server:${VERSION} .
	@$(ECHO) "Built successfully docker image mock-server:${VERSION}"
	@rm -rf $(OUTPUT_DIR)/mock-server-image

test: | pre
	@$(ECHO) "Building test..."
	@$(MAKE) -C $(ROOT_DIR)/test all
	@$(MKDIR) $(OUTPUT_DIR)/test
	@$(CP) -R $(ROOT_DIR)/test/build/* $(OUTPUT_DIR)/test/ 2>/dev/null || true
	@$(ECHO) "Built successfully test"

all: backend application adminclient file relay front tools scripts bintools support-files test

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
	@$(CD) $(ROOT_DIR) && GOGC=40 $(if $(strip $(GOMEMLIMIT)),GOMEMLIMIT=$(GOMEMLIMIT),) GOTOOLCHAIN=$(GO) golangci-lint run --config $(ROOT_DIR)/.golangci.yml --path-prefix $(ROOT_DIR) --timeout $(GOLANGCI_LINT_TIMEOUT)
	@$(ECHO) "Linting completed"

helm:
	@command -v $(HELM) >/dev/null || { $(ECHO) "Error: helm is required"; exit 1; }
	@test -f "$(ROOT_DIR)/install/helm/bk-nodemgr/Chart.lock" || { $(ECHO) "Error: Chart.lock is required; run helm dependency update after changing dependencies"; exit 1; }
	@$(RM) -rf "$(OUTPUT_DIR)/helm/bk-nodemgr"
	@$(MKDIR) "$(OUTPUT_DIR)/helm/bk-nodemgr"
	@$(CP) -R "$(ROOT_DIR)/install/helm/bk-nodemgr/." "$(OUTPUT_DIR)/helm/bk-nodemgr/"
	@$(HELM) lint "$(OUTPUT_DIR)/helm/bk-nodemgr"
	@$(HELM) package "$(OUTPUT_DIR)/helm/bk-nodemgr" --destination "$(OUTPUT_DIR)/helm"
	@$(ECHO) "Built successfully Helm chart: $(OUTPUT_DIR)/helm"
