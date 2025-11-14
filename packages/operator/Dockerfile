# Build the manager binary
FROM registry.redhat.io/ubi9/go-toolset@sha256:380d6de9bbc5a42ca13d425be99958fb397317664bb8a00e49d464e62cc8566c AS builder
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

FROM registry.redhat.io/ubi9/ubi-minimal@sha256:2f06ae0e6d3d9c4f610d32c480338eef474867f435d8d28625f2985e8acde6e8
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
