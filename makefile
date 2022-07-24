NAME := gsnes
BINDIR := ./build
VERSION := $(shell git describe --tags 2>/dev/null)
LDFLAGS := -X 'main.version=$(VERSION) -s -w'

.PHONY: build-darwin
build:
	@go build -trimpath -o $(BINDIR)/darwin-amd64/$(NAME) -ldflags "$(LDFLAGS)" ./cmd/

.PHONY: build-linux
build-linux:
	@GOOS=linux GOARCH=amd64 go build -trimpath -o $(BINDIR)/linux-amd64/$(NAME) -ldflags "$(LDFLAGS)" ./cmd/

.PHONY: build-windows
build-windows:
	@GOOS=windows GOARCH=amd64 go build -trimpath -o $(BINDIR)/windows-amd64/$(NAME).exe -ldflags "$(LDFLAGS)" ./cmd/

.PHONY: clean
clean:
	@-rm -rf $(BINDIR)

.PHONY: help
help:
	@make2help $(MAKEFILE_LIST)
