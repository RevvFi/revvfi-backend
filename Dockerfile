# syntax=docker/dockerfile:1
#
# Multi-stage build producing two separate, minimal runtime images
# (api, indexer) from one shared compile stage. Build one or the other with:
#
#   docker build --target api      -t revvfi-api      .
#   docker build --target indexer  -t revvfi-indexer  .
#
# Why multi-stage: the Go toolchain, module cache, and full source tree are
# only needed to *produce* a binary, never to *run* one. Shipping them in the
# final image (like the old single-stage Dockerfile did with `golang:1.22-alpine`
# as the runtime base) means every container carries a full compiler + package
# cache it never uses at runtime — bigger pulls, bigger attack surface, slower
# `docker compose pull` on EC2. Splitting into stages lets the final image
# contain only the compiled binary and the runtime assets it actually reads.

# ---------------------------------------------------------------------------
# Stage: builder — compiles both binaries. Nothing in this stage ships.
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

# Matches go.mod's `go 1.25.0` — keeping these in sync avoids the toolchain
# silently downloading a second Go version mid-build.

WORKDIR /src

# Alpine's base image doesn't include git, which `go mod download` needs for
# any dependency fetched directly from a VCS URL rather than a proxy.
RUN apk add --no-cache git

# Copy only the dependency manifests first, then download. Docker caches each
# layer by the hash of its inputs — as long as go.mod/go.sum don't change,
# this layer (and the downloaded module cache) is reused across builds even
# when application source changes on every commit. This is the single
# biggest build-time optimization available here: without this split,
# editing one .go file busts the dependency-download cache too.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source and build.
COPY . .

# CGO_ENABLED=0: produces a fully static binary with no libc dependency, so
# it can run on a `FROM scratch`/distroless base with zero shared libraries.
# go-ethereum (a dependency here) builds fine pure-Go without cgo.
#
# -trimpath: strips local filesystem paths (e.g. /src) from the compiled
# binary, so build-machine paths don't leak into stack traces or the binary
# itself — also makes builds reproducible across machines.
#
# -ldflags="-s -w": strips the symbol table and DWARF debug info. This is a
# pure size optimization (typically 20-30% smaller binaries); it does not
# affect runtime behavior. Skip this if you ever need `delve`/gdb debugging
# directly against a container binary.
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/revvfi-api ./cmd/api
RUN go build -trimpath -ldflags="-s -w" -o /out/revvfi-indexer ./cmd/indexer

# ---------------------------------------------------------------------------
# Stage: api — minimal runtime image for the HTTP API server.
# ---------------------------------------------------------------------------
# distroless/static: no shell, no package manager, no coreutils — just glibc-
# free static-binary support plus CA certificates (needed here because both
# the Supabase connection, sslmode=require, and the RPC_URL over https need
# to validate TLS certs) and /etc/passwd entries for a non-root user.
# Smaller and meaningfully more secure than alpine: there's no shell for an
# attacker to get a foothold with even if the app were compromised. The
# tradeoff: you cannot `docker exec -it ... sh` into this container to poke
# around — debug via logs (`docker compose logs api`) instead.
FROM gcr.io/distroless/static-debian12:nonroot AS api

WORKDIR /app

# The indexer's ABI decoder (internal/indexer/decoder) reads
# <ARTIFACT_PATH>/<Contract>.sol/<Contract>.json at startup. The API binary
# doesn't currently use this, but shipping it here too keeps both images
# built the same way and avoids surprises if that changes.
COPY --from=builder /out/revvfi-api ./revvfi-api
COPY --from=builder /src/internal/blockchain/abis ./internal/blockchain/abis

# distroless's "nonroot" variant already runs as uid 65532 by default —
# no explicit USER directive needed (and there's no `adduser` here to make
# one even if we wanted a different uid, since there's no shell/package
# manager in this image).

ENV ARTIFACT_PATH=/app/internal/blockchain/abis
ENV API_PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/revvfi-api"]

# ---------------------------------------------------------------------------
# Stage: indexer — minimal runtime image for the background indexer worker.
# ---------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS indexer

WORKDIR /app

COPY --from=builder /out/revvfi-indexer ./revvfi-indexer
COPY --from=builder /src/internal/blockchain/abis ./internal/blockchain/abis

ENV ARTIFACT_PATH=/app/internal/blockchain/abis

# No EXPOSE: the indexer is a background worker, it doesn't serve HTTP.

ENTRYPOINT ["/app/revvfi-indexer"]
