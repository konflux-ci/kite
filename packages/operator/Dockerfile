# Build the manager binary
FROM registry.redhat.io/ubi9/go-toolset@sha256:bacdac6f6205a52ce6cf5d7b884fbedafa4ba341ab377f5c2c62015022fdf8c7 AS builder
ARG TARGETOS
ARG TARGETARCH
ENV GOTOOLCHAIN=auto

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go

# Not needed (initially using Tekton PipelineRuns resources).
# Uncomment when needed.
#COPY api/ api/
#
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

FROM registry.redhat.io/ubi9/ubi-minimal@sha256:53ea1f6d835898acda5becdb3f8b1292038a480384bbcf994fc0bcf1f7e8eaf7
WORKDIR /

# Copy binaries from buidler
COPY --from=builder /opt/app-root/src .

LABEL name="Konflux KITE Bridge Operator"
LABEL description="Konflux Kite Bridge Operator"
LABEL io.k8s.description="Konflux Kite Bridge Operator"
LABEL io.k8s.display-name="konflux-kite-bridge-operator"
LABEL summary="Konflux Kite Bridge Operator"
LABEL com.redhat.component="kite-bridge-operator"

USER 65532:65532

ENTRYPOINT ["/manager"]
