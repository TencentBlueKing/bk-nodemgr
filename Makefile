.PHONY: tidy build test pre backend application file relay front docker-build-server all clean doc tools script_tools

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

backend: pre
	@$(ECHO) "Building backend $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-backend $(ROOT_DIR)/cmd/backend/main.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-backend"

application: pre
	@$(ECHO) "Building application $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-application $(ROOT_DIR)/cmd/application/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-application"

file: pre
	@$(ECHO) "Building file $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-file $(ROOT_DIR)/cmd/file/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-file"

relay: pre
	@$(ECHO) "Building proxy $(VERSION)..."
	CGO_ENABLED=0 $(GO) build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodemgr-relay $(ROOT_DIR)/cmd/relay/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodemgr-relay"

front: pre
	@$(ECHO) "Building frontend..."
	@$(CD) $(ROOT_DIR)/front && $(NPM) i && $(NPM) build
	@$(CP) -R $(ROOT_DIR)/front/dist $(OUTPUT_DIR)/
	@$(ECHO) "Built successfully: frontend"

tools: pre
	@$(ECHO) "Building tools..."
	@$(MAKE) -C $(ROOT_DIR)/tools platform-builds -e UPX_ENABLED=1

	$(MKDIR) $(OUTPUT_DIR)/tools
	@$(CP) -r $(ROOT_DIR)/tools/build/$(VERSION)/* $(OUTPUT_DIR)/tools
	@$(ECHO) "Built successfully tools"

script_tools: pre
	@$(ECHO) "Building script tools..."

	@$(MKDIR) $(OUTPUT_DIR)/script_tools/bintool

	@$(MKDIR) $(OUTPUT_DIR)/script_tools/bintool/agent_linux_amd64
	@$(CP) $(ROOT_DIR)/script_tools/gsectl/agent/linux/gsectl $(OUTPUT_DIR)/script_tools/bintool/agent_linux_amd64

	@$(MKDIR) $(OUTPUT_DIR)/script_tools/bintool/agent_linux_arm64
	@$(CP) $(ROOT_DIR)/script_tools/gsectl/agent/linux/gsectl $(OUTPUT_DIR)/script_tools/bintool/agent_linux_arm64

	@$(MKDIR) $(OUTPUT_DIR)/script_tools/bintool/agent_darwin_amd64
	@$(CP) $(ROOT_DIR)/script_tools/gsectl/agent/darwin/gsectl $(OUTPUT_DIR)/script_tools/bintool/agent_darwin_amd64

	@$(MKDIR) $(OUTPUT_DIR)/script_tools/bintool/agent_windows_amd64
	@$(CP) $(ROOT_DIR)/script_tools/gsectl/agent/windows/gsectl.bat $(OUTPUT_DIR)/script_tools/bintool/agent_windows_amd64

	@$(MKDIR) $(OUTPUT_DIR)/script_tools/bintool/proxy_linux_amd64
	@$(CP) $(ROOT_DIR)/script_tools/gsectl/proxy/linux/gsectl $(OUTPUT_DIR)/script_tools/bintool/proxy_linux_amd64

	@$(MKDIR) $(OUTPUT_DIR)/script_tools/bintool/proxy_linux_arm64
	@$(CP) $(ROOT_DIR)/script_tools/gsectl/proxy/linux/gsectl $(OUTPUT_DIR)/script_tools/bintool/proxy_linux_arm64

	@$(CD) $(OUTPUT_DIR)/script_tools/ && $(TAR) bintool.tgz bintool/

	@$(ECHO) "Built successfully $(OUTPUT_DIR)/script_tools/bintool.tgz"
	@$(ECHO) "Built successfully script tools"

docker-build-server: backend application file front tools
	@$(ECHO) "Building docker images..."
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr/Dockerfile $(OUTPUT_DIR)
	@$(CP) $(ROOT_DIR)/install/docker-compose/serviced.sh $(OUTPUT_DIR)
	@$(CD) $(OUTPUT_DIR) && docker build -t bk-nodemgr-server:v${VERSION} .
	@$(ECHO) "Built successfully docker images bk-nodemgr-server:v${VERSION}"

docker-build-apigw-sync: pre
	@$(ECHO) "Building docker image bk-nodemgr-apigw-sync..."
	@$(MKDIR) $(OUTPUT_DIR)/apigw-sync
	@$(CP) -R $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/support-files $(OUTPUT_DIR)/apigw-sync/
	@$(CP) -R $(ROOT_DIR)/docs/apigw/* $(OUTPUT_DIR)/apigw-sync/support-files
	@$(CP) $(ROOT_DIR)/install/images/bk-nodemgr-apigw-sync/Dockerfile $(OUTPUT_DIR)/apigw-sync
	@$(CD) $(OUTPUT_DIR)/apigw-sync && docker build -t bk-nodemgr-apigw-sync:v${VERSION} .
	@$(ECHO) "Built successfully docker image bk-nodemgr-apigw-sync:v${VERSION}"

all: backend application file relay front tools script_tools

clean:
	@$(ECHO) "Cleaning build directory..."
	@$(RM) -rf build
	$(MAKE) -C $(ROOT_DIR)/tools clean
	@$(ECHO) "Cleaned build directory"

doc:
	@$(ECHO) "Open http://localhost:6060 to view the documentation"
	godoc -http=localhost:6060
