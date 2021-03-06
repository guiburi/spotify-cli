BINDIR := $(shell pwd)/bin

.PHONY: install_deps
install_deps:
	go mod download

.PHONY: build
build: install_deps
	mkdir -p $(BINDIR)
	packr build -o $(BINDIR)/spotify-cli ./cmd/spotify-cli

.PHONY: clean
clean:
	rm -rf $(BINDIR)

.PHONY: test
test:
	go test -v ./...

.PHONY: images
images:
	plantuml -tpng img/components.puml
	plantuml -tpng img/workflow.puml

.DEFAULT_GOAL := build
