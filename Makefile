GO_PACKAGES=$(shell go list ./...)
GO ?= $(shell command -v go 2> /dev/null)
BUILD_HASH ?= $(shell git rev-parse HEAD)
BUILD_VERSION ?= $(shell git ls-remote --tags --refs git://github.com/mattermost/mmctl | tail -n1 | sed 's/.*\///')
# Needed to avoid install shadow in brew which is not permitted
ADVANCED_VET ?= TRUE
ENTERPRISE_DIR ?= ${MM_SERVER_PATH}/../enterprise
VENDOR_MM_SERVER_DIR ?= vendor/github.com/mattermost/mattermost-server/v5/
ENTERPRISE_HASH ?= $(shell cat enterprise_hash)
TESTFLAGS = -mod=vendor -timeout 30m -race -v
LDFLAGS += -X "github.com/mattermost/mmctl/commands.BuildHash=$(BUILD_HASH)"
BUILD_TAGS =

.PHONY: all
all: build

-include config.override.mk
include config.mk

# Prepares the enterprise build if exists. The IGNORE stuff is a hack to get the Makefile to execute the commands outside a target
ifneq ($(wildcard ${ENTERPRISE_DIR}/.*),)
	TESTFLAGS += -ldflags '-X "github.com/mattermost/mmctl/commands.EnableEnterpriseTests=true" -X "github.com/mattermost/mattermost-server/v5/model.BuildEnterpriseReady=true"'
	BUILD_TAGS +=enterprise
	IGNORE:=$(shell echo Enterprise build selected, preparing)
	IGNORE:=$(shell rm -rf $(VENDOR_MM_SERVER_DIR)/enterprise)
	IGNORE:=$(shell cp -R $(ENTERPRISE_DIR) $(VENDOR_MM_SERVER_DIR))
	IGNORE:=$(shell git -C $(VENDOR_MM_SERVER_DIR)/enterprise checkout $(ENTERPRISE_HASH) --quiet)
	IGNORE:=$(shell rm -f $(VENDOR_MM_SERVER_DIR)/imports/imports.go)
	IGNORE:=$(shell cp $(VENDOR_MM_SERVER_DIR)/enterprise/imports/imports.go $(VENDOR_MM_SERVER_DIR)/imports/)
endif

.PHONY: build
build: vendor check
	go build -ldflags '$(LDFLAGS)' -mod=vendor
	md5sum < mmctl | cut -d ' ' -f 1 > mmctl.md5.txt

.PHONY: install
install: vendor check
	go install -ldflags '$(LDFLAGS)' -mod=vendor

.PHONY: package
package: vendor
	mkdir -p build

	@echo Build Linux amd64
	env GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -mod=vendor
	tar cf build/linux_amd64.tar mmctl
	md5sum < build/linux_amd64.tar | cut -d ' ' -f 1 > build/linux_amd64.tar.md5.txt

	@echo Build OSX amd64
	env GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -mod=vendor
	tar cf build/darwin_amd64.tar mmctl
	md5sum < build/darwin_amd64.tar | cut -d ' ' -f 1 > build/darwin_amd64.tar.md5.txt

	@echo Build Windows amd64
	env GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -mod=vendor
	zip build/windows_amd64.zip mmctl.exe
	md5sum < build/windows_amd64.zip | cut -d ' ' -f 1 > build/windows_amd64.zip.md5.txt

	rm mmctl mmctl.exe

.PHONY: gofmt
gofmt:
	@echo Running gofmt
	@for package in $(GO_PACKAGES); do \
		echo "Checking "$$package; \
		files=$$(go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' $$package); \
		if [ "$$files" ]; then \
			gofmt_output=$$(gofmt -d -s $$files 2>&1); \
			if [ "$$gofmt_output" ]; then \
				echo "$$gofmt_output"; \
				echo "Gofmt failure"; \
				exit 1; \
			fi; \
		fi; \
	done
	@echo Gofmt success

.PHONY: govet
govet:
ifeq ($(ADVANCED_VET), TRUE)
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \

	@echo Running golangci-lint
	golangci-lint run ./...
	@if ! [ -x "$$(command -v mattermost-govet)" ]; then \
		echo "mattermost-govet is not installed. Please install it executing \"GO111MODULE=off go get -u github.com/mattermost/mattermost-govet\""; \
		exit 1; \
	fi;
	@echo Running mattermost-govet
	$(GO) vet -vettool=$(GOPATH)/bin/mattermost-govet -license -structuredLogging -inconsistentReceiverName -tFatal -equalLenAsserts ./...
endif
	@echo Govet success

.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit:
	@echo Running unit tests
	$(GO) test $(TESTFLAGS) -tags 'unit $(BUILD_TAGS)' $(GO_PACKAGES)

.PHONY: test-e2e
test-e2e:
	@echo Running e2e tests
	MM_SERVER_PATH=${MM_SERVER_PATH} $(GO) test $(TESTFLAGS) -tags 'e2e $(BUILD_TAGS)' $(GO_PACKAGES)

.PHONY: test-all
test-all:
	@echo Running all tests
	MM_SERVER_PATH=${MM_SERVER_PATH} $(GO) test $(TESTFLAGS) -tags 'unit e2e $(BUILD_TAGS)' $(GO_PACKAGES)

.PHONY: coverage
coverage:
	MM_SERVER_PATH=${MM_SERVER_PATH} $(GO) test $(TESTFLAGS) -tags 'unit e2e $(BUILD_TAGS)' -coverprofile=coverage.txt ./...
	$(GO) tool cover -html=coverage.txt

.PHONY: check
check: gofmt govet

.PHONY: vendor
vendor:
	go mod vendor
	go mod tidy

.PHONY: mocks
mocks:
	mockgen -destination=mocks/client_mock.go -copyright_file=mocks/copyright.txt -package=mocks github.com/mattermost/mmctl/client Client

.PHONY: docs
docs:
	rm -rf docs
	go run -mod=vendor mmctl.go docs
