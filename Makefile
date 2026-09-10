BINARY := dailyup
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)

install:
	go install $(LDFLAGS) .
