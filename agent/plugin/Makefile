# SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause)

OUTPUT := .output
CLANG ?= clang
LIBBPF_SRC := $(abspath ./libbpf-bootstrap/libbpf/src)
BPFTOOL_SRC := $(abspath ./libbpf-bootstrap/bpftool/src)
LIBBPF_OBJ := $(abspath $(OUTPUT)/libbpf.a)
BPFTOOL_OUTPUT ?= $(abspath $(OUTPUT)/bpftool)
BPFTOOL ?= $(BPFTOOL_OUTPUT)/bootstrap/bpftool

# Determine architecture based on system's machine type
ARCH ?= $(shell uname -m | sed 's/x86_64/x86/' \
			 | sed 's/arm.*/arm/' \
			 | sed 's/aarch64/arm64/' \
			 | sed 's/ppc64le/powerpc/' \
			 | sed 's/mips.*/mips/' \
			 | sed 's/riscv64/riscv/' \
			 | sed 's/loongarch64/loongarch/')
VMLINUX := ./vmlinux.h

INCLUDES := -I$(OUTPUT) -I../../libbpf/include/uapi -I$(dir $(VMLINUX))
CFLAGS := -g -Wall -Wno-unknown-pragmas
ALL_LDFLAGS := $(LDFLAGS) $(EXTRA_LDFLAGS)

#kernel version
VERSION_INFO := $(shell uname -r | cut -d'-' -f1)
VERSION_MAJOR := $(shell echo $(VERSION_INFO) | cut -d'.' -f1)
VERSION_MINOR := $(shell echo $(VERSION_INFO) | cut -d'.' -f2)
VERSION_PATCH := $(shell echo $(VERSION_INFO) | cut -d'.' -f3)
export VERSION_INFO
export VERSION_MAJOR
export VERSION_MINOR
export VERSION_PATCH

# User-space applications
APPS = probe

# Get system include paths for clang
CLANG_BPF_SYS_INCLUDES = $(shell $(CLANG) -v -E - </dev/null 2>&1 \
	| sed -n '/<...> search starts here:/,/End of search list./{ s| \(/.*\)|-idirafter \1|p }')
# Output message settings for make
ifeq ($(V),1)
	Q =
	msg =
else
	Q = @
	msg = @printf '  %-8s %s%s\n'					\
		      "$(1)"						\
		      "$(patsubst $(abspath $(OUTPUT))/%,%,$(2))"	\
		      "$(if $(3), $(3))";
	MAKEFLAGS += --no-print-directory
endif

define allow-override
  $(if $(or $(findstring environment,$(origin $(1))),\
            $(findstring command line,$(origin $(1)))),,\
    $(eval $(1) = $(2)))
endef

$(call allow-override,CC,$(CROSS_COMPILE)cc)
$(call allow-override,LD,$(CROSS_COMPILE)ld)

# Check if bpftool is installed, otherwise install it
.PHONY: all
all: dependence $(APPS)
.PHONY: dependence
dependence:check_bpftool check_clang 
	@echo "Kernel version is $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)"
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > $(VMLINUX)

.PHONY: check_bpftool
check_bpftool:
	@if ! command -v bpftool &> /dev/null; then \
		echo "bpftool Not installed, installing..."; \
		sudo yum update; \
		sudo yum install -y linux-tools-$(shell uname -r); \
	else \
		echo "bpftool Have been installed"; \
	fi
.PHONY: check_clang

check_clang:
	@if ! command -v clang &> /dev/null; then \
		echo "clang Not installed, installing..."; \
		sudo yum update; \
		sudo yum install -y clang; \
	else \
		echo "clang Have been installed"; \
	fi

.PHONY: clean
clean:
	$(call msg,CLEAN)
	$(Q)rm -rf $(OUTPUT) $(APPS)
	
.PHONY: clean2
clean2:
	$(call msg,CLEAN)
	rm -f $(APPS)
	rm -f $(OUTPUT)/*.skel.h
	rm -f $(OUTPUT)/*.o

$(OUTPUT) $(OUTPUT)/libbpf $(BPFTOOL_OUTPUT):
	$(call msg,MKDIR,$@)
	$(Q)mkdir -p $@

# Build libbpf
$(LIBBPF_OBJ): $(wildcard $(LIBBPF_SRC)/*.[ch] $(LIBBPF_SRC)/Makefile) | $(OUTPUT)/libbpf
	$(call msg,LIB,$@)
	$(Q)$(MAKE) -C $(LIBBPF_SRC) BUILD_STATIC_ONLY=1		      \
		    OBJDIR=$(dir $@)/libbpf DESTDIR=$(dir $@)		      \
		    INCLUDEDIR= LIBDIR= UAPIDIR=			      \
		    install

# Build bpftool
$(BPFTOOL): | $(BPFTOOL_OUTPUT)
	$(call msg,BPFTOOL,$@)
	$(Q)$(MAKE) ARCH= CROSS_COMPILE= OUTPUT=$(BPFTOOL_OUTPUT)/ -C $(BPFTOOL_SRC) bootstrap
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > $(VMLINUX)

# Build BPF code
$(OUTPUT)/%.bpf.o: %.bpf.c $(LIBBPF_OBJ) $(wildcard %.h) $(VMLINUX) | $(OUTPUT) $(BPFTOOL)
	$(call msg,BPF,$@)
	$(Q)$(CLANG) -g -O2 -target bpf -D__TARGET_ARCH_$(ARCH)		      \
		     $(INCLUDES) $(CLANG_BPF_SYS_INCLUDES)		      \
		     -c $(filter %.c,$^) -o $(patsubst %.bpf.o,%.tmp.bpf.o,$@)
	$(Q)$(BPFTOOL) gen object $@ $(patsubst %.bpf.o,%.tmp.bpf.o,$@)

# Generate BPF skeletons
$(OUTPUT)/%.skel.h: $(OUTPUT)/%.bpf.o | $(OUTPUT) $(BPFTOOL)
	$(call msg,GEN-SKEL,$@)
	$(Q)$(BPFTOOL) gen skeleton $< > $@

# Build user-space code
$(patsubst %,$(OUTPUT)/%.o,$(APPS)): %.o: %.skel.h

$(OUTPUT)/%.o: %.c $(wildcard %.h) | $(OUTPUT)
	$(call msg,CC,$@)
	$(Q)$(CC) $(CFLAGS) $(INCLUDES) -c $(filter %.c,$^) -o $@

# Build application binary
$(APPS): %: $(OUTPUT)/%.o $(LIBBPF_OBJ) | $(OUTPUT)
	$(call msg,BINARY,$@)
	$(Q)$(CC) $(CFLAGS) $^ $(ALL_LDFLAGS) -lelf -lz -o $@

# delete failed targets
.DELETE_ON_ERROR:

# keep intermediate (.skel.h, .bpf.o, etc) targets
.SECONDARY:
