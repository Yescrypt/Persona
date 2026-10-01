BINARY  := persona
PKG     := ./cmd/persona
VERSION ?= 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)
PREFIX  ?= $(HOME)/.local

.PHONY: build install test fmt vet clean tidy

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

install: build
	mkdir -p $(PREFIX)/bin
	cp $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "Installed to $(PREFIX)/bin/$(BINARY)"

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BINARY) wordlist.txt usernames.txt
