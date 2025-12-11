.PHONY: all build clean plugin_bintool_v2 plugin_bintool_v3

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

all: plugin_bintool_v2 plugin_bintool_v3
	@$(CD) $(OUTPUT_DIR)/ && $(TAR) plugin_bintool.tgz plugin_bintool/
	@$(ECHO) "Built successfully $(OUTPUT_DIR)/plugin_bintool.tgz"
	@$(ECHO) "Built successfully plugin bintool"

plugin_bintool_v2:
	@$(ECHO) "Building plugin bintool v2..."

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v2/linux_amd64
	@$(CP) $(CURRENT_DIR)/v2/linux/* $(OUTPUT_DIR)/plugin_bintool/v2/linux_amd64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v2/linux_arm64
	@$(CP) $(CURRENT_DIR)/v2/linux/* $(OUTPUT_DIR)/plugin_bintool/v2/linux_arm64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v2/darwin_amd64
	@$(CP) $(CURRENT_DIR)/v2/darwin/* $(OUTPUT_DIR)/plugin_bintool/v2/darwin_amd64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v2/windows_amd64
	@$(CP) $(CURRENT_DIR)/v2/windows/* $(OUTPUT_DIR)/plugin_bintool/v2/windows_amd64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v2/aix6_ppc64
	@$(CP) $(CURRENT_DIR)/v2/aix/* $(OUTPUT_DIR)/plugin_bintool/v2/aix6_ppc64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v2/aix7_ppc64
	@$(CP) $(CURRENT_DIR)/v2/aix/* $(OUTPUT_DIR)/plugin_bintool/v2/aix7_ppc64

plugin_bintool_v3:
	@$(ECHO) "Building v3 plugin bintool v3..."
	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v3/linux_amd64
	@$(CP) $(CURRENT_DIR)/v3/linux/* $(OUTPUT_DIR)/plugin_bintool/v3/linux_amd64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v3/linux_arm64
	@$(CP) $(CURRENT_DIR)/v3/linux/* $(OUTPUT_DIR)/plugin_bintool/v3/linux_arm64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v3/darwin_amd64
	@$(CP) $(CURRENT_DIR)/v3/darwin/* $(OUTPUT_DIR)/plugin_bintool/v3/darwin_amd64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v3/windows_amd64
	@$(CP) $(CURRENT_DIR)/v3/windows/* $(OUTPUT_DIR)/plugin_bintool/v3/windows_amd64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v3/aix6_ppc64
	@$(CP) $(CURRENT_DIR)/v3/aix/* $(OUTPUT_DIR)/plugin_bintool/v3/aix6_ppc64

	@$(MKDIR) $(OUTPUT_DIR)/plugin_bintool/v3/aix7_ppc64
	@$(CP) $(CURRENT_DIR)/v3/aix/* $(OUTPUT_DIR)/plugin_bintool/v3/aix7_ppc64

clean:
	@$(ECHO) "Cleaning build directory..."
	@$(RM) -rf build
	@$(ECHO) "Cleaned build directory"