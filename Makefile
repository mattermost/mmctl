GO_PACKAGES=$(shell go list ./...)

all: build

build: vendor check
	go build -mod=vendor

install: vendor check
	go install -mod=vendor

fmt:
	go fmt $(GO_PACKAGES)

vet:
	go vet $(GO_PACKAGES)

check: fmt vet

vendor:
	go mod vendor
