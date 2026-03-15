.PHONY: build test lint clean run bench release-dry release docker-build docker-test

GOTAGS := -tags sqlite_fts5

build:
	go build $(GOTAGS) -o bin/engram ./cmd/engram

test:
	go test $(GOTAGS) ./...

lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -rf bin/

run:
	go run $(GOTAGS) ./cmd/engram

bench:
	go test $(GOTAGS) -bench=. -benchmem -timeout 300s ./benchmarks/...

release-dry:
	goreleaser release --snapshot --clean --skip=publish

release:
	goreleaser release --clean

docker-build:
	docker build -t engram:latest .

docker-test: docker-build
	docker run --rm -v $(PWD):/workspace engram:latest status
