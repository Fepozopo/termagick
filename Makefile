# Makefile - helper targets for building `termagick` Linux binaries and images using Docker BuildKit / buildx
#
# Usage examples (from repository root):
#   make -f termagick/Makefile build            # build the runtime image (final stage)
#   make -f termagick/Makefile binary           # extract a linux/amd64 binary to ./out
#   make -f termagick/Makefile out/termagick    # same as `binary` (phony target produces ./out/termagick)
#   make -f termagick/Makefile multiarch        # build & push multi-arch image (requires registry auth)
#   make -f termagick/Makefile clean            # remove local artifacts
#
# Notes:
#  - This Makefile uses Docker BuildKit / buildx. Ensure Docker CLI has buildx enabled:
#      `docker buildx version`
#  - The repo Dockerfile (multi-stage) must be at `Dockerfile` relative to the current working directory.
#  - The builder stage produces the binary at /out/termagick in the builder image.
#  - You can override variables on the command line, e.g.:
#      make -f termagick/Makefile TARGET=linux/arm64 binary
#

# Tooling
DOCKER ?= docker
BUILDX ?= $(DOCKER) buildx
DOCKER_BUILDKIT ?= 1

# Paths / names
DOCKERFILE ?= Dockerfile
OUTDIR ?= out
BINARY ?= termagick

# Image names / tags
IMAGE ?= termagick:latest
BUILDER_IMAGE ?= termagick-builder:local
GOLANG_IMAGE ?= golang:1.25

# Platforms (comma-separated for buildx)
PLATFORMS ?= linux/amd64,linux/arm64
TARGET ?= linux/amd64

# Buildx builder (optional name used when creating a builder)
BUILDX_BUILDER_NAME ?= termagick-builder

# Speed / caching knobs (user can override)
CACHE_DIR ?= /tmp/termagick-docker-cache

.PHONY: help build builder image binary out multiarch push create-builder rm-builder clean docker-build

help:
	@printf "Makefile targets (use with -f termagick/Makefile):\n\n"
	@printf "  build        Build the final runtime image (uses Docker BuildKit).\n"
	@printf "  binary       Produce a linux binary in ./out (uses buildx --output local).\n"
	@printf "  docker-build Produce a linux binary in ./out using Docker.\n"
	@printf "  out/$(BINARY)  Alias for 'binary' (creates $(OUTDIR)/$(BINARY)).\n"
	@printf "  multiarch    Build (and optionally push) a multi-arch image via buildx --platform=$(PLATFORMS).\n"
	@printf "  push         Push an already-built image tag to registry (simple docker push).\n"
	@printf "  create-builder  Create a dedicated buildx builder (idempotent).\n"
	@printf "  rm-builder      Remove the named buildx builder.\n"
	@printf "  clean        Remove $(OUTDIR) and local cache artifacts.\n"
	@printf "\nVariables you can override: IMAGE, BINARY, OUTDIR, PLATFORMS, TARGET, DOCKERFILE\n\n"

# Build the final runtime image (final stage of Dockerfile)
build:
	@echo "Building runtime image: $(IMAGE)"
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(BUILDX) build \
		--platform $(TARGET) \
		-f $(DOCKERFILE) \
		--build-arg GOLANG_IMAGE=$(GOLANG_IMAGE) \
		-t $(IMAGE) \
		.

# Build only the builder stage and extract the binary to ./out using buildx --output local
binary: $(OUTDIR)/$(BINARY)

$(OUTDIR)/$(BINARY):
	@echo "Building builder stage and exporting binary to $(OUTDIR)/$(BINARY) (platform: $(TARGET))"
	@mkdir -p $(OUTDIR)
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(BUILDX) build \
		--platform $(TARGET) \
		-f $(DOCKERFILE) \
		--build-arg GOLANG_IMAGE=$(GOLANG_IMAGE) \
		--target builder \
		--output type=local,dest=$(OUTDIR) \
		-t $(BUILDER_IMAGE) \
		.

	@if [ -f "$(OUTDIR)/$(BINARY)" ]; then \
		chmod +x "$(OUTDIR)/$(BINARY)"; \
		echo "Binary exported to $(OUTDIR)/$(BINARY)"; \
	else \
		echo "ERROR: expected $(OUTDIR)/$(BINARY) but it was not created"; exit 1; \
	fi

docker-build:
	@echo "Building builder image and extracting binary by docker create/cp (GOLANG_IMAGE=$(GOLANG_IMAGE), TARGET=$(TARGET))"
	# Use buildx to build the builder stage for the requested platform and load it into the local docker daemon.
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(BUILDX) build \
		--platform $(TARGET) \
		--build-arg GOLANG_IMAGE=$(GOLANG_IMAGE) \
		-f $(DOCKERFILE) \
		--target builder \
		-t $(BUILDER_IMAGE) \
		--load \
		.
	@mkdir -p $(OUTDIR)
	CONTAINER_ID=`$(DOCKER) create $(BUILDER_IMAGE)` && \
	 $(DOCKER) cp $$CONTAINER_ID:/out/$(BINARY) $(OUTDIR)/$(BINARY) && \
	 $(DOCKER) rm -v $$CONTAINER_ID; \
	 chmod +x $(OUTDIR)/$(BINARY); \
	 echo "Binary extracted to $(OUTDIR)/$(BINARY)"


# Build multi-arch image and push to registry (requires `docker login` beforehand).
# This will push images for all platforms in $(PLATFORMS) to the tag $(IMAGE).
multiarch:
	@echo "Building and pushing multi-arch image: $(IMAGE) (platforms: $(PLATFORMS))"
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(BUILDX) build \
		--platform $(PLATFORMS) \
		-f $(DOCKERFILE) \
		--build-arg GOLANG_IMAGE=$(GOLANG_IMAGE) \
		-t $(IMAGE) \
		--push \
		.

# If you want to build multi-arch but only produce local images (without pushing),
# remove `--push` and add `--load` (note: --load supports only single-platform).
# push: push a local image tag using plain docker push
push:
	@echo "Pushing image: $(IMAGE)"
	$(DOCKER) push $(IMAGE)

# Create a dedicated buildx builder (idempotent)
create-builder:
	@echo "Creating buildx builder named '$(BUILDX_BUILDER_NAME)' (if not exists)"
	@$(BUILDX) inspect $(BUILDX_BUILDER_NAME) >/dev/null 2>&1 || $(BUILDX) create --name $(BUILDX_BUILDER_NAME) --use
	@echo "Using buildx builder: $$( $(BUILDX) inspect --bootstrap | head -n1 )"

# Remove the named builder (if you created one and want to delete it)
rm-builder:
	@echo "Removing buildx builder named '$(BUILDX_BUILDER_NAME)' (if exists)"
	-@$(BUILDX) rm $(BUILDX_BUILDER_NAME) || true

clean:
	@echo "Cleaning $(OUTDIR) and caches"
	-@rm -rf $(OUTDIR)
	@echo "Done."

.PHONY: all
all: build