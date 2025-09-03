# Multi-stage Dockerfile to build a Linux binary of `termagick` that depends on ImageMagick.
# Usage (from repository root or the `termagick` directory):
#  - Simple build:
#      docker build -f termagick/Dockerfile -t termagick:latest .
#  - Build for a specific target (using build args):
#      docker build -f termagick/Dockerfile --build-arg TARGETOS=linux --build-arg TARGETARCH=amd64 -t termagick:linux .
#  - Recommended: use BuildKit / buildx for multi-arch builds:
#      DOCKER_BUILDKIT=1 docker buildx build --platform linux/amd64 -f termagick/Dockerfile -t termagick:linux .
#
# Notes:
#  - This project uses `gographics/imagick` which requires CGO and ImageMagick development headers at build time
#    and the ImageMagick runtime libraries at container runtime.
#  - The builder stage installs build dependencies (gcc, libmagickwand-dev) and produces the binary.
#  - The final stage is a slim Debian image that installs the ImageMagick runtime packages and copies the binary.

# ----------------------------
# Builder stage
# ----------------------------
ARG GOLANG_IMAGE=golang:1.25
FROM ${GOLANG_IMAGE} AS builder

# Build args to control target OS/ARCH. Defaults produce a linux/amd64 binary.
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG CGO_ENABLED=1

# Ensure apt is noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies needed to compile ImageMagick-using code.
# - git: for go modules that may require fetching non-module sources
# - gcc, build-essential: for cgo compilation
# - pkg-config: helps discover ImageMagick libs
# - libmagickwand-dev: ImageMagick development headers for linking at build time
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    gcc \
    make \
    pkg-config \
    libmagickwand-dev \
    && rm -rf /var/lib/apt/lists/*

# Create working directory and ensure go modules are downloaded first (cache-friendly)
WORKDIR /src

# Copy go.mod and go.sum first to leverage Docker cache for dependencies
COPY go.mod go.sum ./

# Download modules (cached layer)
RUN go mod download

# Copy the rest of the project
COPY . .

# Set build environment for cross-building and CGO
ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# Allow CGO to use the Xpreprocessor flag (needed for some architectures)
ENV CGO_CFLAGS_ALLOW='-Xpreprocessor'

# Output dir for the built binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    mkdir -p /out && \
    go build -ldflags="-s -w" -o /out/termagick ./...

# ----------------------------
# Runtime stage
# ----------------------------
FROM debian:bookworm-slim AS runtime

# Install ImageMagick runtime libraries. `imagemagick` will pull the runtime libs needed by the binary.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    ca-certificates \
    imagemagick \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user to run the binary (optional but recommended)
RUN groupadd -r app && useradd -r -g app -d /home/app -s /sbin/nologin app \
    && mkdir -p /home/app \
    && chown -R app:app /home/app

# Copy the built binary from the builder stage
COPY --from=builder /out/termagick /usr/local/bin/termagick

# Ensure binary is executable
RUN chmod +x /usr/local/bin/termagick

USER app
WORKDIR /home/app

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/termagick"]
# Default command (shows help if no args are provided)
CMD ["--help"]
