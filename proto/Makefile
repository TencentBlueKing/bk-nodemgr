.PHONY: default all clean

# directories
ROOT_DIR = $(CURDIR)/..

# application v3
APPLICATION_V3_PROTO_DIR = $(CURDIR)/application/api/v3
APPLICATION_V3_GO_OUT_DIR = $(ROOT_DIR)/pkg/proto/application/api/v3
APPLICATION_V3_PROTO_FILES = $(wildcard $(APPLICATION_V3_PROTO_DIR)/*.proto)
APPLICATION_V3_GO_FILES = $(patsubst $(APPLICATION_V3_PROTO_DIR)/%.proto,$(APPLICATION_V3_GO_OUT_DIR)/%.pb.go,$(APPLICATION_V3_PROTO_FILES))

# backend v3
BACKEND_V3_PROTO_DIR = $(CURDIR)/backend/api/v3
BACKEND_V3_GO_OUT_DIR = $(ROOT_DIR)/pkg/proto/backend/api/v3
BACKEND_V3_PROTO_FILES = $(wildcard $(BACKEND_V3_PROTO_DIR)/*.proto)
BACKEND_V3_GO_FILES = $(patsubst $(BACKEND_V3_PROTO_DIR)/%.proto,$(BACKEND_V3_GO_OUT_DIR)/%.pb.go,$(BACKEND_V3_PROTO_FILES))

# cmd
MKDIR  = mkdir -p
ECHO   = $(if $(filter Linux,$(shell uname)),@echo -e,@echo)
CD     = cd
CP     = cp
RM     = rm
SH     = sh
PROTOC = protoc -I.
FORMAT = clang-format -i

default: all
pre:
	$(MKDIR) $(APPLICATION_V3_GO_OUT_DIR)
	$(MKDIR) $(BACKEND_V3_GO_OUT_DIR)

	$(FORMAT) $(APPLICATION_V3_PROTO_FILES)
	$(FORMAT) $(BACKEND_V3_PROTO_FILES)

$(APPLICATION_V3_GO_OUT_DIR)/%.pb.go: $(APPLICATION_V3_PROTO_DIR)/%.proto
	$(PROTOC) -I=$(APPLICATION_V3_PROTO_DIR) --go_out=paths=source_relative:$(APPLICATION_V3_GO_OUT_DIR) $<

application-v3: pre
	$(MAKE) $(APPLICATION_V3_GO_FILES)
	$(ECHO) "successfully generated application api/v3"

$(BACKEND_V3_GO_OUT_DIR)/%.pb.go: $(BACKEND_V3_PROTO_DIR)/%.proto
	$(PROTOC) -I=$(BACKEND_V3_PROTO_DIR) --go_out=paths=source_relative:$(BACKEND_V3_GO_OUT_DIR) $<

backend-v3: pre
	$(MAKE) $(BACKEND_V3_GO_FILES)
	$(ECHO) "successfully generated backend api/v3"

all: application-v3 backend-v3
	$(ECHO) "all generating done"

clean:
	$(RM) -f $(APPLICATION_V3_GO_FILES)
	$(RM) -f $(BACKEND_V3_GO_FILES)
