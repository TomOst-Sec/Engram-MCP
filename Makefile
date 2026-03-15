.PHONY: build test lint clean run

TAGS := sqlite_fts5

build:
	go build -tags "$(TAGS)" -o bin/engram ./cmd/engram

test:
	go test -tags "$(TAGS)" ./...

lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -rf bin/

run:
	go run -tags "$(TAGS)" ./cmd/engram
