# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -tags sqlite_fts5 \
    -ldflags="-s -w" \
    -o /engram ./cmd/engram

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache sqlite-libs git ca-certificates

COPY --from=builder /engram /usr/local/bin/engram

# Create non-root user
RUN adduser -D -h /home/engram engram
USER engram
WORKDIR /workspace

# Default: start MCP server
ENTRYPOINT ["engram"]
CMD ["serve"]
