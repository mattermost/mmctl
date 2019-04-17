GO_PACKAGES=$(shell go list ./...)

all: build

build: vendor check
	go build -mod=vendor

install: vendor check
	go install -mod=vendor

package: vendor check
	mkdir -p build

	@echo Build Linux amd64
	env GOOS=linux GOARCH=amd64 go build -mod=vendor
	tar cf build/linux_amd64.tar mmctl

	@echo Build OSX amd64
	env GOOS=darwin GOARCH=amd64 go build -mod=vendor
	tar cf build/darwin_amd64.tar mmctl

	@echo Build Windows amd64
	env GOOS=windows GOARCH=amd64 go build -mod=vendor
	zip build/windows_amd64.zip mmctl.exe

	rm mmctl mmctl.exe

fmt:
	go fmt $(GO_PACKAGES)

vet:
	go vet $(GO_PACKAGES)

check: fmt vet

vendor:
	go mod vendor
