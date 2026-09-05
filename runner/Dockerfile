# syntax=docker/dockerfile:1

# The Go version the image runs submissions with. It must be >= the `go`
# directive of every challenge's `go.mod` (currently 1.24) or those challenges
# will not build. See packages/runner/README.md.
ARG GO_VERSION=1.25

# --- entrypoint -------------------------------------------------------------
# Built statically so the runtime stage needs no libc beyond what the Go
# toolchain itself brings.
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gofinity-runner .

# --- pre-warmed build cache -------------------------------------------------
# Compiling the standard library from cold costs more than a submission's entire
# time budget. This stage compiles and tests a representative module so the
# image ships a populated GOCACHE; each run copies it onto its tmpfs, because
# the container's root filesystem is read-only. See cache.go.
FROM golang:${GO_VERSION}-alpine AS prewarm
ENV CGO_ENABLED=0 GOCACHE=/opt/gocache GOFLAGS=-mod=mod GOTOOLCHAIN=local GOPROXY=off
WORKDIR /prewarm
COPY prewarm/ ./
RUN go test ./... && go build ./... && chmod -R a+rX /opt/gocache

# --- runtime ----------------------------------------------------------------
# The Go toolchain has to be present to compile the submission, so this is the
# same base rather than a scratch image. Nothing else is added to it: no shell
# tooling beyond the busybox that comes with alpine, no package installs, no
# network client.
FROM golang:${GO_VERSION}-alpine AS runtime

# A fixed, non-root uid so the API can pass `--user 65532:65532` without
# depending on what the image happens to define.
RUN addgroup -g 65532 -S runner && adduser -u 65532 -S -G runner -h /home/runner runner

COPY --from=build /out/gofinity-runner /usr/local/bin/gofinity-runner
COPY --from=prewarm /opt/gocache /opt/gocache

# Everything writable lives under /tmp, which the API mounts as a tmpfs
# (`exec` included - `go test` builds a binary there and runs it). The rest of
# the filesystem is expected to be read-only at runtime.
ENV HOME=/tmp \
    GOFINITY_WORKDIR=/tmp/work \
    GOFINITY_PREWARMED_CACHE=/opt/gocache \
    GOCACHE=/tmp/gocache \
    GOMODCACHE=/tmp/gomodcache \
    GOPATH=/tmp/gopath \
    GOENV=off \
    GOFLAGS=-mod=mod \
    GOTOOLCHAIN=local \
    GOPROXY=off \
    GOSUMDB=off \
    CGO_ENABLED=0 \
    GOMAXPROCS=2

USER 65532:65532
WORKDIR /tmp

# No CMD: the payload arrives in GOFINITY_PAYLOAD or on stdin, never as args.
ENTRYPOINT ["/usr/local/bin/gofinity-runner"]
