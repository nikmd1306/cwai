BINARY := cwai
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/nikmd1306/cwai/cmd.Version=$(VERSION)

.PHONY: build install clean test vet lint check eval eval-compare eval-test

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

eval:
	go build -tags eval -ldflags "$(LDFLAGS)" -o $(BINARY) . && ./$(BINARY) eval --dataset experiments/dataset/golden.jsonl --out experiments/runs

eval-compare:
	go build -tags eval -ldflags "$(LDFLAGS)" -o $(BINARY) . && ./$(BINARY) eval compare $(RUN_A) $(RUN_B)

eval-test:
	go test -tags eval ./internal/eval/... -v
