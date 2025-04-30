# Image information
IMAGE_NAME := ghcr.io/jonathaninc/dev/backend
IMAGE_TAG := latest
FULL_IMAGE_NAME := $(IMAGE_NAME):$(IMAGE_TAG)

# Docker build arguments
DOCKER_BUILD_ARGS := 

.PHONY: all build push clean

all: build push

# Build the Docker image
build:
	@echo "Building Docker image: $(FULL_IMAGE_NAME)"
	docker build $(DOCKER_BUILD_ARGS) -t $(FULL_IMAGE_NAME) .

# Push the Docker image to GitHub Container Registry
push:
	@echo "Pushing image to GitHub Container Registry: $(FULL_IMAGE_NAME)"
	docker push $(FULL_IMAGE_NAME)

# Clean local Docker images
clean:
	@echo "Removing local Docker image: $(FULL_IMAGE_NAME)"
	-docker rmi $(FULL_IMAGE_NAME)
