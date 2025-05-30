.PHONY: tidy build test pre backend application file relay front docker-build all clean doc tools

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

# cmd
MKDIR = mkdir -p
ECHO  = $(if $(filter Linux,$(shell uname)),echo -e,echo)
CD    = cd
CP    = cp
RM    = rm
SH    = sh
NPM   = pnpm

default: all

pre:
	@$(MKDIR) $(OUTPUT_DIR)
	go mod tidy

backend: pre
	@$(ECHO) "Building backend $(VERSION)..."
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-backend $(ROOT_DIR)/cmd/backend/main.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodeman-backend"

application: pre
	@$(ECHO) "Building application $(VERSION)..."
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-application $(ROOT_DIR)/cmd/application/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodeman-application"

file: pre
	@$(ECHO) "Building file $(VERSION)..."
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-file $(ROOT_DIR)/cmd/file/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodeman-file"

relay: pre
	@$(ECHO) "Building proxy $(VERSION)..."
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-relay $(ROOT_DIR)/cmd/relay/*.go
	@$(ECHO) "Built successfully: $(OUTPUT_DIR)/bk-nodeman-relay"

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

docker-build: backend application file front tools
	@$(ECHO) "Building docker images..."
	@$(CP) $(ROOT_DIR)/install/images/Dockerfile $(OUTPUT_DIR)
	@$(CD) $(OUTPUT_DIR) && docker build -t bk-nodeman:v${VERSION} .
	@$(ECHO) "Built successfully docker images bk-nodeman:v${VERSION}"


all: backend application file relay front tools

clean:
	@$(ECHO) "Cleaning build directory..."
	@$(RM) -rf build
	$(MAKE) -C $(ROOT_DIR)/tools clean
	@$(ECHO) "Cleaned build directory"

doc:
	@$(ECHO) "Open http://localhost:6060 to view the documentation"
	godoc -http=localhost:6060
