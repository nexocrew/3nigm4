# Makefile for 3nigm4 project
# 	
# Targets:
# 	all: Builds the code
# 	build: Builds the code
# 	fmt: Formats the source files
# 	clean: cleans the code
# 	install: Installs the code to the GOPATH
# 	iref: Installs referenced projects
#	test: Runs the tests
#

# Go parameters
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test -v
#GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w

# the following folder is used for third party
# dependencies (see http://peter.bourgon.org/go-in-production/ 
# for tecnique description).
GOPATH := $(GOPATH):$(CURDIR)/_vendor

# test with benchmarking switch
TEST_BENCHMARK ?= no

# Package lists
TOPLEVEL_PKG := github.com/nexocrew/3nigm4
IMPL_LIST := authserver storageservice 3n4cli ishtmservice ishtmdispatcher #<-- Implementation directories
COMMON_LIST := lib/version lib/logo lib/itm lib/logger lib/crypto \
	lib/messages lib/client lib/filemanager lib/s3 lib/auth \
	lib/storageclient lib/ishtm/will lib/ishtm/commons lib/ishtm/db \
	lib/sender lib/sender/smtp

# List building
ALL_LIST = $(IMPL_LIST) $(COMMON_LIST)

BUILD_LIST = $(foreach int, $(ALL_LIST), $(int)_build)
CLEAN_LIST = $(foreach int, $(ALL_LIST), $(int)_clean)
INSTALL_LIST = $(foreach int, $(ALL_LIST), $(int)_install)
IREF_LIST = $(foreach int, $(ALL_LIST), $(int)_iref)
TEST_LIST = $(foreach int, $(ALL_LIST), $(int)_test)
FMT_TEST = $(foreach int, $(ALL_LIST), $(int)_fmt)

# All are .PHONY for now because dependencyness is hard
.PHONY: $(CLEAN_LIST) $(TEST_LIST) $(FMT_LIST) $(INSTALL_LIST) $(BUILD_LIST) $(IREF_LIST)

all: install
clean: $(CLEAN_LIST)
install: $(INSTALL_LIST)
test: $(TEST_LIST)
iref: $(IREF_LIST)
fmt: $(FMT_TEST)

$(CLEAN_LIST): %_clean:
	$(GOCLEAN) $(TOPLEVEL_PKG)/$*
$(INSTALL_LIST): %_install:
	$(GOINSTALL) $(TOPLEVEL_PKG)/$*
$(IREF_LIST): %_iref:
	$(GODEP) $(TOPLEVEL_PKG)/$*
$(TEST_LIST): %_test:
ifeq ($(TEST_BENCHMARK), yes)
	$(GOTEST) --bench=. $(TOPLEVEL_PKG)/$*
else
	$(GOTEST) $(TOPLEVEL_PKG)/$*
endif
$(FMT_TEST): %_fmt:
	$(GOFMT) ./$*
