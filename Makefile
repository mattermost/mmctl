GO_PACKAGES=$(shell go list ./...)

all: build

build:
	go build

fmt:
	go fmt $(GO_PACKAGES)

vet:
	go vet $(GO_PACKAGES)

check: fmt vet

vendor:
	go mod vendor
