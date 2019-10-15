GOPATH=$(shell pwd)/vendor:$(shell pwd)
GOBIN=$(shell pwd)/bin
GOFILES=$(wildcard /src*.go)
GONAME=passcheck

build:
	@echo "Building $(GOFILES) to $(GOBIN)"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o bin/$(GONAME) $(GOFILES)

install:
	@GOPATH=