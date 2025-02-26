.PHONY: tidy build test pre backend application front docker-build all clean doc

# directories
ROOT_DIR = $(CURDIR)
OUTPUT_DIR = $(ROOT_DIR)/build/$(VERSION)

# version
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITTAG    = $(shell git describe --tags --always)
GITHASH   = $(shell git rev-parse --short HEAD)
VERSION  ?= ${GITTAG}-$(shell date +%y.%m.%d)

# ldflags
# output directory for release package and version for command line
LDVersionFLAG = "-X github.com/TencentBlueKing/bk-nodemgr/internal/version.VERSION=${VERSION} \
    	-X github.com/TencentBlueKing/bk-nodemgr/internal/version.BUILDTIME=${BUILDTIME} \
    	-X github.com/TencentBlueKing/bk-nodemgr/internal/version.GITHASH=${GITHASH}"


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
	$(MKDIR) $(OUTPUT_DIR)
	go mod tidy

backend: pre
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-backend $(ROOT_DIR)/cmd/backend/main.go

application: pre
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-application $(ROOT_DIR)/cmd/application/*.go

file: pre
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-file $(ROOT_DIR)/cmd/file/*.go

front: pre
	$(CD) $(ROOT_DIR)/front && $(NPM) i && $(NPM) build
	$(CP) -R $(ROOT_DIR)/front/dist $(OUTPUT_DIR)/

docker-build: backend application file front
	$(CP) $(ROOT_DIR)/install/images/Dockerfile $(OUTPUT_DIR)
	$(CD) $(OUTPUT_DIR) && docker build -t bk-nodeman:v${VERSION} .

all: backend application file front

clean:
	$(RM) -rf build

doc:
	godoc -http=localhost:6060