# ── Stage 1: Build ────────────────────────────────────────────────────────────
# Use the full Go image only to compile — it's large (~800 MB)
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency manifests first — Docker layer caches this if go.mod/go.sum
# haven't changed, so `go mod download` only runs when deps actually change.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /go-http-server ./cmd/server

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
# Distroless: no shell, no package manager — minimal attack surface (~5 MB)
FROM gcr.io/distroless/static-debian12

COPY --from=builder /go-http-server /go-http-server

# Never run as root
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/go-http-server"]
