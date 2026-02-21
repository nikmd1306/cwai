BINARY := cwai
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install clean test vet lint check

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f $(BINARY)

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

check: build vet lint test
