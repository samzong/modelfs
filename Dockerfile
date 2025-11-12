# Dockerfile for GoReleaser builds
# GoReleaser builds binaries separately and places them in linux/${TARGETARCH}/manager
# This Dockerfile only copies the pre-built binary for optimal image size and build speed
FROM gcr.io/distroless/static:nonroot
ARG TARGETPLATFORM
ARG TARGETARCH

WORKDIR /app

# Copy pre-built binary from GoReleaser build context
# GoReleaser ensures linux/${TARGETARCH}/manager exists before building
COPY linux/${TARGETARCH}/manager /app/manager

USER 65532:65532

ENTRYPOINT ["/app/manager"]

