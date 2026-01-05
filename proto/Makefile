.PHONY: default all clean

# directories
ROOT_DIR = $(CURDIR)/..

# application v3
APPLICATION_V3_PROTO_DIR = $(CURDIR)/application/api/v3
APPLICATION_V3_GO_OUT_DIR = $(ROOT_DIR)/pkg/proto/application/api/v3
APPLICATION_V3_SWAGGER_OUT_DIR = $(ROOT_DIR)/docs/api/swagger/application/api/v3
APPLICATION_V3_PROTO_FILES = $(wildcard $(APPLICATION_V3_PROTO_DIR)/*.proto)
APPLICATION_V3_GO_FILES = $(patsubst $(APPLICATION_V3_PROTO_DIR)/%.proto,$(APPLICATION_V3_GO_OUT_DIR)/%.pb.go,$(APPLICATION_V3_PROTO_FILES))

# backend v3
BACKEND_V3_PROTO_DIR = $(CURDIR)/backend/api/v3
BACKEND_V3_GO_OUT_DIR = $(ROOT_DIR)/pkg/proto/backend/api/v3
BACKEND_V3_SWAGGER_OUT_DIR = $(ROOT_DIR)/docs/api/swagger/backend/api/v3
BACKEND_V3_PROTO_FILES = $(wildcard $(BACKEND_V3_PROTO_DIR)/*.proto)
BACKEND_V3_GO_FILES = $(patsubst $(BACKEND_V3_PROTO_DIR)/%.proto,$(BACKEND_V3_GO_OUT_DIR)/%.pb.go,$(BACKEND_V3_PROTO_FILES))

# file v3
FILE_V3_PROTO_DIR = $(CURDIR)/file/api/v3
FILE_V3_GO_OUT_DIR = $(ROOT_DIR)/pkg/proto/file/api/v3
FILE_V3_SWAGGER_OUT_DIR = $(ROOT_DIR)/docs/api/swagger/file/api/v3
FILE_V3_PROTO_FILES = $(wildcard $(FILE_V3_PROTO_DIR)/*.proto)
FILE_V3_GO_FILES = $(patsubst $(FILE_V3_PROTO_DIR)/%.proto,$(FILE_V3_GO_OUT_DIR)/%.pb.go,$(FILE_V3_PROTO_FILES))

# backend callback
BACKEND_CALLBACK_PROTO_DIR = $(CURDIR)/backend/callback
BACKEND_CALLBACK_GO_OUT_DIR = $(ROOT_DIR)/pkg/proto/backend/callback
BACKEND_CALLBACK_SWAGGER_OUT_DIR = $(ROOT_DIR)/docs/api/swagger/backend/callback
BACKEND_CALLBACK_PROTO_FILES = $(wildcard $(BACKEND_CALLBACK_PROTO_DIR)/*.proto)
BACKEND_CALLBACK_GO_FILES = $(patsubst $(BACKEND_CALLBACK_PROTO_DIR)/%.proto,$(BACKEND_CALLBACK_GO_OUT_DIR)/%.pb.go,$(BACKEND_CALLBACK_PROTO_FILES))


# cmd
MKDIR   = mkdir -p
ECHO    = $(if $(filter Linux,$(shell uname)),@echo -e,@echo)
CD      = cd
CP      = cp
RM      = rm
SH      = sh
PROTOC  = protoc -Ithirdparty/
FORMAT  = clang-format -i
REPLACE = sed -i 's/,omitempty//g'
FIXSWAGGER = ./convert_swagger_type.sh

default: all
pre:
	$(MKDIR) $(APPLICATION_V3_GO_OUT_DIR) $(APPLICATION_V3_SWAGGER_OUT_DIR)
	$(MKDIR) $(BACKEND_V3_GO_OUT_DIR) $(BACKEND_V3_SWAGGER_OUT_DIR)
	$(MKDIR) $(BACKEND_CALLBACK_GO_OUT_DIR) $(BACKEND_CALLBACK_SWAGGER_OUT_DIR)
	$(MKDIR) $(FILE_V3_GO_OUT_DIR) $(FILE_V3_SWAGGER_OUT_DIR)

	$(FORMAT) $(APPLICATION_V3_PROTO_FILES)
	$(FORMAT) $(BACKEND_V3_PROTO_FILES)
	$(FORMAT) $(BACKEND_CALLBACK_PROTO_FILES)
	$(FORMAT) $(FILE_V3_PROTO_FILES)

$(APPLICATION_V3_GO_OUT_DIR)/%.pb.go: $(APPLICATION_V3_PROTO_DIR)/%.proto
	$(PROTOC) -I=$(APPLICATION_V3_PROTO_DIR) --go_out=paths=source_relative:$(APPLICATION_V3_GO_OUT_DIR) --openapiv2_out=$(APPLICATION_V3_SWAGGER_OUT_DIR) --openapiv2_opt output_format=json --openapiv2_opt=json_names_for_fields=false $<
	$(REPLACE) $@

application-v3: pre
	$(MAKE) $(APPLICATION_V3_GO_FILES)
	$(ECHO) "successfully generated application api/v3"

$(BACKEND_V3_GO_OUT_DIR)/%.pb.go: $(BACKEND_V3_PROTO_DIR)/%.proto
	$(PROTOC) -I=$(BACKEND_V3_PROTO_DIR) --go_out=paths=source_relative:$(BACKEND_V3_GO_OUT_DIR) --openapiv2_out=$(BACKEND_V3_SWAGGER_OUT_DIR) --openapiv2_opt output_format=json --openapiv2_opt=json_names_for_fields=false $<
	$(REPLACE) $@

backend-v3: pre
	$(MAKE) $(BACKEND_V3_GO_FILES)
	$(ECHO) "successfully generated backend api/v3"

$(BACKEND_CALLBACK_GO_OUT_DIR)/%.pb.go: $(BACKEND_CALLBACK_PROTO_DIR)/%.proto
	$(PROTOC) -I=$(BACKEND_CALLBACK_PROTO_DIR) --go_out=paths=source_relative:$(BACKEND_CALLBACK_GO_OUT_DIR) --openapiv2_out=$(BACKEND_CALLBACK_SWAGGER_OUT_DIR) --openapiv2_opt output_format=json --openapiv2_opt=json_names_for_fields=false $<
	$(REPLACE) $@

backend-callback: pre
	$(MAKE) $(BACKEND_CALLBACK_GO_FILES)
	$(ECHO) "successfully generated backend callback"

$(FILE_V3_GO_OUT_DIR)/%.pb.go: $(FILE_V3_PROTO_DIR)/%.proto
	$(PROTOC) -I=$(FILE_V3_PROTO_DIR) --go_out=paths=source_relative:$(FILE_V3_GO_OUT_DIR) --openapiv2_out=$(FILE_V3_SWAGGER_OUT_DIR) --openapiv2_opt output_format=json --openapiv2_opt=json_names_for_fields=false $<
	$(REPLACE) $@

file-v3: pre
	$(MAKE) $(FILE_V3_GO_FILES)
	$(ECHO) "successfully generated file api/v3"

all: application-v3 backend-v3 file-v3 backend-callback
	$(FIXSWAGGER) $(ROOT_DIR)/docs/api/swagger
	$(ECHO) "convert_swagger_type done"
	$(ECHO) "all generating done"

clean:
	$(RM) -f $(APPLICATION_V3_GO_FILES)
	$(RM) -f $(BACKEND_V3_GO_FILES)
	$(RM) -f $(BACKEND_CALLBACK_GO_FILES)
	$(RM) -f $(FILE_V3_GO_FILES)
