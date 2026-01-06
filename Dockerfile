# syntax = docker/dockerfile:1.4.1
# Copyright 2021-present The Atlas Authors. All rights reserved.
# This source code is licensed under the Apache 2.0 license found
# in the LICENSE file in the root directory of this source tree.

# Build the Atlas CLI binary
FROM golang:1.24-alpine3.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the go source
COPY cmd/ cmd/
COPY internal/ internal/
COPY schemahcl/ schemahcl/
COPY sdk/ sdk/
COPY sql/ sql/

# Build from cmd/atlas
WORKDIR /workspace/cmd/atlas
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} CGO_ENABLED=0 \
    go build -o /atlas -a .

FROM alpine:3.20
WORKDIR /
COPY --from=builder --chmod=755 /atlas /usr/local/bin/atlas
USER 65532:65532

LABEL org.opencontainers.image.source="https://github.com/cracicorp/atlas"
LABEL org.opencontainers.image.description="Atlas CLI - Database schema migration tool"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/usr/local/bin/atlas"]
