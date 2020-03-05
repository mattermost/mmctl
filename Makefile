.PHONY: vendor docs mocks
GO_PACKAGES=$(shell go list ./...)
GO ?= $(shell command -v go 2> /dev/null)
BUILD_HASH ?= $(shell git rev-parse HEAD)
BUILD_VERSION ?= $(shell git ls-remote --tags --refs git://github.com/mattermost/mmctl | tail -n1 | sed 's/.*\///')
# Needed to avoid install shadow in brew which is not permitted
ADVANCED_VET ?= TRUE

LDFLAGS += -X "github.com/mattermost/mmctl/commands.BuildHash=$(BUILD_HASH)"

all: build

build: vendor check
	go build -ldflags '$(LDFLAGS)' -mod=vendor
	md5sum < mmctl | cut -d ' ' -f 1 > mmctl.md5.txt

install: vendor check
	go install -ldflags '$(LDFLAGS)' -mod=vendor

package: vendor check
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

test: test-unit

test-unit:
	@echo Running unit tests
	$(GO) test -mod=vendor -race -v -tags unit $(GO_PACKAGES)

test-e2e:
	@echo Running e2e tests
	$(GO) test -mod=vendor -race -v -tags e2e $(GO_PACKAGES)

test-all:
	@echo Running all tests
	$(GO) test -mod=vendor -race -v -tags 'unit e2e' $(GO_PACKAGES)

.PHONY: coverage
coverage:
	$(GO) test -mod=vendor -race -tags unit -coverprofile=coverage.txt ./...
	$(GO) tool cover -html=coverage.txt

check: gofmt govet

vendor:
	go mod vendor
	go mod tidy

mocks:
	mockgen -destination=mocks/client_mock.go -copyright_file=mocks/copyright.txt -package=mocks github.com/mattermost/mmctl/client Client

docs:
	rm -rf docs
	go run -mod=vendor mmctl.go docs
