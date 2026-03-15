.PHONY: build test lint clean run

build:
	go build -o bin/engram ./cmd/engram

test:
	go test ./...

lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -rf bin/

run:
	go run ./cmd/engram
