.PHONY: all build clean bintool

# version
GITTAG  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
VERSION ?= ${GITTAG}-$(shell date +%y.%m.%d)

# directories
CURRENT_DIR := $(CURDIR)
OUTPUT_DIR  := $(CURDIR)/build/$(VERSION)

# cmd
MKDIR = mkdir -p
ECHO  = $(if $(filter Linux,$(shell uname)),echo -e,echo)
CD    = cd
CP    = cp
RM    = rm
SH    = sh
TAR   = tar -zcf

default: all

all: bintool

bintool:
	@$(ECHO) "Building gsectl bintool..."
	@$(MKDIR) $(OUTPUT_DIR)/bintool

	@$(MKDIR) $(OUTPUT_DIR)/bintool/agent_linux_amd64
	@$(CP) $(CURRENT_DIR)/agent/linux/gsectl $(OUTPUT_DIR)/bintool/agent_linux_amd64

	@$(MKDIR) $(OUTPUT_DIR)/bintool/agent_linux_arm64
	@$(CP) $(CURRENT_DIR)/agent/linux/gsectl $(OUTPUT_DIR)/bintool/agent_linux_arm64

	@$(MKDIR) $(OUTPUT_DIR)/bintool/agent_darwin_amd64
	@$(CP) $(CURRENT_DIR)/agent/darwin/gsectl $(OUTPUT_DIR)/bintool/agent_darwin_amd64

	@$(MKDIR) $(OUTPUT_DIR)/bintool/agent_windows_amd64
	@$(CP) $(CURRENT_DIR)/agent/windows/gsectl.bat $(OUTPUT_DIR)/bintool/agent_windows_amd64

	@$(MKDIR) $(OUTPUT_DIR)/bintool/proxy_linux_amd64
	@$(CP) $(CURRENT_DIR)/proxy/linux/gsectl $(OUTPUT_DIR)/bintool/proxy_linux_amd64

	@$(MKDIR) $(OUTPUT_DIR)/bintool/proxy_linux_arm64
	@$(CP) $(CURRENT_DIR)/proxy/linux/gsectl $(OUTPUT_DIR)/bintool/proxy_linux_arm64

	@$(CD) $(OUTPUT_DIR)/ && $(TAR) bintool.tgz bintool/

	@$(ECHO) "Built successfully $(OUTPUT_DIR)/bintool.tgz"
	@$(ECHO) "Built successfully bintool"

clean:
	@$(ECHO) "Cleaning build directory..."
	@$(RM) -rf build
	@$(ECHO) "Cleaned build directory"