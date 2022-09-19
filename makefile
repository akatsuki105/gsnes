NAME := gsnes
BINDIR := ./build
VERSION := $(shell git describe --tags 2>/dev/null)
LDFLAGS := -X 'main.version=$(VERSION) -s -w'

## Build native
build:
	@go build -trimpath -o $(BINDIR)/$(NAME) -ldflags "$(LDFLAGS)" ./cmd/

## Build for profiler(for development)
build-profiler:
	@go build -o $(BINDIR)/profiler/profiler ./profiler

## Test sfc core
test:
	@go run ./tester/PeterLemon 

## Clean repository
clean:
	@-rm -rf $(BINDIR)

## Show help
help:
	@make2help $(MAKEFILE_LIST)

## Run godoc on localhost:3000
doc:
	@godoc -http=:3000

.PHONY: build build-profiler test clean help doc
