# syntax=docker/dockerfile:1.7

# 1) Build statically
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
WORKDIR /src

# Speed up builds with module & build caches
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    apk add --no-cache git ca-certificates

# Copy mod files first for better caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy rest of the source
COPY . .

# Build static binary (no CGO), strip symbols, trim paths
ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -buildvcs=false -o /out/app .

# 2) Minimal runtime (nonroot)
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# Your server expects these env vars; set sensible defaults
ENV CSP_PORT=8080
# IMPORTANT: you MUST override this in production
ENV CSP_SEED=PLEASE_SET_ME

# Copy the binary
COPY --from=build /out/app /app/server

# Expose the HTTP port (matches default CSP_PORT)
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
