.PHONY: build test lint clean run bench release-dry release

export CGO_CFLAGS := -DSQLITE_ENABLE_FTS5
export CGO_LDFLAGS := -lm

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

bench:
	go test -bench=. -benchmem -timeout 300s ./benchmarks/...

release-dry:
	goreleaser release --snapshot --clean --skip=publish

release:
	goreleaser release --clean
